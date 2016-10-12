package mongodb

import (
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	mongodbURL = "mongodb://192.168.1.210/account"
)

func TestDial(t *testing.T) {
	db, err := Dial(mongodbURL, 128)
	if err != nil {
		t.Error("mongodb: cannot connect to %v, err: %v", mongodbURL, err)
	}
	db.DBAction("account", "account", func(c *mgo.Collection) error {

		err := c.Insert(bson.M{"name": "good"})
		if err != nil {
			t.Error(err)
		}

		count, err := c.Find(nil).Count()
		if err != nil {
			t.Error(err)
		}
		t.Log("count:", count)

		_, err = c.RemoveAll(nil)
		if err != nil {
			t.Error(err)
		}

		return err
	})
}
