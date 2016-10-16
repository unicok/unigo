package services

import "testing"

// func TestService(t *testing.T) {
// 	Init()
// 	spew.Dump(defaultPool)
// 	if conn, _ := GetService("/backends/snowflake"); conn == nil {
// 		t.Log("get service failed")
// 	} else {
// 		t.Log("get service successed")
// 	}

// 	if GetServieWithID("/backends/snowflake", "snowflake1") == nil {
// 		t.Log("get service with id failed")
// 	} else {
// 		t.Log("get service with id successed")
// 	}
// }

func TestGetServiceAddress(t *testing.T) {
	addr, err := GetServiceAddress("consul.service.consul")
	if err != nil {
		t.Error(err)
	}
	t.Log(addr)
}
