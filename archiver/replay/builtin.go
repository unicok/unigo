package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/yuin/gopher-lua"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Change is a data change
type Change struct {
	Collection string
	Field      string
	Doc        interface{}
}

//RedoRecord is a redo record represents complete transaction
type RedoRecord struct {
	API     string
	UID     int32
	TS      uint64
	Changes []Change
}

func (p *ToolBox) builtinHelp(L *lua.LState) int {
	fmt.Println(`
    REDO Replay Tool
    Commands:

    > help()                                    -- print this text
    > print(redo:lenght())                      -- print redolog length
    > print(redo:get(1))                        -- print a document
    > redo:mgo("mongodb://172.17.42.1/mydb")    -- attach to mongodb
    > redo:replay(1)                            -- replay redolog#1
    > dofile("/go/scripts/json.lua")            -- require scripts
    > tbl = decode(redo:get(1))                 -- convert json to table
    > print(tbl.TS)                             -- print TS
    `)
	return 0
}

func (p *ToolBox) bGet(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.([]rec); ok {
		if L.GetTop() == 2 {
			idx := L.CheckInt(2) - 1
			if idx >= 0 && idx < len(v) {
				elem := v[idx]
				r := p.read(idx, elem.dbIdx, elem.key)
				if r != nil {
					r.TS >>= 22 // keep only millisecond part
				}
				bin, _ := json.MarshalIndent(r, "", "\t")
				L.Push(lua.LString(bin))
				return 1
			}
			L.ArgError(1, "index out of range")
		}
	}
	L.ArgError(1, "invalid userdata")
	return 0
}

func (p *ToolBox) bLength(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.([]rec); ok {
		L.Push(lua.LNumber(len(v)))
		return 1
	}
	L.ArgError(1, "invalid userdata")
	return 0
}

func (p *ToolBox) bMgo(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if _, ok := ud.Value.([]rec); ok {
		if L.GetTop() == 2 {
			url := L.CheckString(2)
			sess, err := mgo.Dial(url)
			if err != nil {
				L.ArgError(1, err.Error())
				return 0
			}
			p.mgo = sess
			p.mgoURL = url
			return 0
		}
		L.Push(lua.LString(p.mgoURL))
		return 1
	}
	L.ArgError(1, "invalid userdata")
	return 0
}

func (p *ToolBox) bReplay(L *lua.LState) int {
	if p.mgo == nil {
		L.Error(lua.LString("use mgo(url) to bind first"), 0)
		return 0
	}

	ud := L.CheckUserData(1)
	if v, ok := ud.Value.([]rec); ok {
		if L.GetTop() == 2 {
			idx := L.CheckInt(2) - 1
			if idx >= 0 && idx < len(v) {
				elem := v[idx]
				r := p.read(idx, elem.dbIdx, elem.key)
				if r != nil {
					L.Push(lua.LBool(doUpdate(r, p.mgo)))
				} else {
					L.Push(lua.LBool(false))
				}
				return 1
			}

			L.ArgError(1, "index out of range")
		}
	}
	L.ArgError(1, "invalid userdata")
	return 0
}

func (p *ToolBox) read(idx int, dbIdx int, key uint64) *RedoRecord {
	var r *RedoRecord
	err := p.dbs[dbIdx].View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltDbBucket))
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, uint64(key))
		bin := b.Get(k)
		if bin == nil {
			return errors.New("record not found")
		}
		r = new(RedoRecord)
		err := bson.Unmarshal(bin, r)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		return nil
	}
	return r
}

func doUpdate(r *RedoRecord, sess *mgo.Session) bool {
	mdb := sess.DB("")
	for k := range r.Changes {
		var err error
		if r.Changes[k].Field != "" {
			_, err = mdb.C(r.Changes[k].Collection).Upsert(bson.M{"userid": r.UID}, bson.M{"$set": bson.M{r.Changes[k].Field: r.Changes[k].Doc}})
		} else {
			_, err = mdb.C(r.Changes[k].Collection).Upsert(bson.M{"userid": r.UID}, r.Changes[k].Doc)
		}
		if err != nil {
			return false
		}
	}
	return true
}
