package main

import (
	"testing"
	"time"

	redo "github.com/unicok/unigo/lib/nsq-redo"
)

type testdoc struct {
	name string
	age  int
}

func TestRedo(t *testing.T) {
	doc := testdoc{}
	r := redo.NewRedoRecord(1, "test1", ts())
	doc.name = "name1"
	doc.age = 18
	r.AddChange("test", "xxx", doc)
	redo.Publish(r)

	r = redo.NewRedoRecord(2, "test2", ts())
	doc.name = "name2"
	doc.age = 22
	r.AddChange("test", "", doc)
	redo.Publish(r)

	time.Sleep(time.Second)
}

const tsMask = 0x1FFFFFFFFFF

func ts() uint64 {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	return (uint64(t) & tsMask) << 22
}
