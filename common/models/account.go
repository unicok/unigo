package models

import "gopkg.in/mgo.v2/bson"

type (
	Account struct {
		Id         bson.ObjectId       `bson:"_id"`
		UserId     uint64              `bson:"user_id"`
		Name       string              `bson:"username"`
		Pass       string              `bson:"password"`
		DeviceName string              `bson:"device_name"`
		DeviceId   string              `bson:"device_id"`
		DeviceType int                 `bson:"device_type"`
		OpenUUID   string              `bson:"open_udid"`
		Lang       string              `bson:"user_lang"`
		LoginIP    string              `bson:"login_ip"`
		LoginTime  bson.MongoTimestamp `bson:"login_time"`
		Token      string              `bson:"login_token"`
	}
)
