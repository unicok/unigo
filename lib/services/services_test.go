package services

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestService(t *testing.T) {
	Init()
	spew.Dump(defaultPool)
	if conn, _ := GetService("/backends/snowflake"); conn == nil {
		t.Log("get service failed")
	} else {
		t.Log("get service successed")
	}

	if GetServieWithID("/backends/snowflake", "snowflake1") == nil {
		t.Log("get service with id failed")
	} else {
		t.Log("get service with id successed")
	}
}
