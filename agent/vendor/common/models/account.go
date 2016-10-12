package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	Account struct {
		Id          bson.ObjectId `bson:"_id"`
		UserId      uint64        `bson:"user_id"`
		UserName    string        `bson:"username"`
		Password    string        `bson:"password"`
		DeviceName  string        `bson:"device_name"`
		DeviceId    string        `bson:"device_id"`
		DeviceType  int32         `bson:"device_type"`
		OpenUUID    string        `bson:"open_udid"`
		Lang        string        `bson:"user_lang"`
		LoginIP     string        `bson:"login_ip"`
		LastLoginAt time.Time     `bson:"last_login_at"`
		Token       string        `bson:"login_token"`
	}
)
