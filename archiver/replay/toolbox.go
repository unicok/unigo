package main

import (
	"encoding/binary"
	"log"
	"path/filepath"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/yuin/gopher-lua"
	"gopkg.in/mgo.v2"
)

const (
	boltDbBucket = "REDOLOG"
	layout       = "2006-01-02"
)

type rec struct {
	dbIdx int    // file
	key   uint64 // key of file
}

// ToolBox is
type ToolBox struct {
	L      *lua.LState // the lua virtual machine
	dbs    []*bolt.DB  // all opened boltdb
	recs   []rec
	mgo    *mgo.Session
	mgoURL string
}

type fileSort []string

func (a fileSort) Len() int {
	return len(a)
}

func (a fileSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a fileSort) Less(i, j int) bool {
	lo := "REDO-2006-01-02.rdo"
	tmA, _ := time.Parse(lo, a[i])
	tmB, _ := time.Parse(lo, a[j])
	return tmA.Unix() < tmB.Unix()
}

func NewToolBox(dir string) *ToolBox {
	t := new(ToolBox)
	// lookup *.rdo
	files, err := filepath.Glob(dir + "/*.rdo")
	if err != nil {
		log.Println(err)
		return nil
	}

	// sort by creation time
	sort.Sort(fileSort(files))

	// open all db
	for _, file := range files {
		db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 2 * time.Second, ReadOnly: true})
		if err != nil {
			log.Println(err)
			continue
		}
		t.dbs = append(t.dbs, db)
	}

	// reindex all keys
	log.Println("loading database")
	for i := range t.dbs {
		t.dbs[i].View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(boltDbBucket))
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				t.recs = append(t.recs, rec{i, binary.BigEndian.Uint64(k)})
			}
			return nil
		})
	}

	// int lua machine
	log.Println("init lua machine")
	t.L = lua.NewState()
	// register
	t.register()
	log.Println("ready")

	t.L.DoString("help()")
	return t
}

// Close the toolbox
func (p *ToolBox) Close() {
	p.L.Close()
	for _, db := range p.dbs {
		db.Close()
	}
	if p.mgo != nil {
		p.mgo.Close()
	}
}

func (p *ToolBox) register() {
	mt := p.L.NewTypeMetatable("mt_reclist")
	p.L.SetGlobal("mt_reclist", mt)
	p.L.SetField(mt, "__index", p.L.SetFuncs(p.L.NewTable(), map[string]lua.LGFunction{
		"get":    p.bGet,
		"length": p.bLength,
		"mgo":    p.bMgo,
		"replay": p.bReplay,
	}))
}

func (p *ToolBox) exec(cmd string) {
	if err := p.L.DoString(cmd); err != nil {
		log.Println(err)
	}
}
