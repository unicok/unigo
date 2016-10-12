package main

import (
	"common/models"
	"lib/proto/snowflake"
	"lib/utils"
	"time"

	"github.com/prometheus/common/log"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func authOrReg(p *models.Account) (*models.Account, error) {
	// 登录验证步骤
	// 1.查询数据库,验证用户名密码 验证成功
	// 2.更新token

	// 验证用户
	var acc = models.Account{}
	err := accDB.DBAction(DBAccount, "account", func(c *mgo.Collection) error {
		return c.Find(p.UserName).One(&acc)
	})
	if err == nil {
		// 验证密码
		if acc.Password != p.Password {
			return nil, ErrPasswordInvalid
		}
	} else if err == mgo.ErrNotFound {
		// 创建ID
		res, err := sfCli.Next(context.Background(), &snowflake.Snowflake_Key{Name: "userid"})
		if err != nil {
			return nil, err
		}

		log.Info("new userid:", res.Value)

		// 创建用户
		acc = *p
		acc.Id = bson.NewObjectId()
		acc.UserId = uint64(res.Value)
		err = accDB.DBAction(DBAccount, "account", func(c *mgo.Collection) error {
			return c.Insert(&acc)
		})
		if err != nil {
			return nil, err
		}

	} else {
		return nil, err
	}

	// 更新Token
	newtoken, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	acc.Token = newtoken
	acc.LastLoginAt = time.Now()
	err = accDB.DBAction(DBAccount, "account", func(c *mgo.Collection) error {
		return c.UpdateId(acc.Id, bson.M{"$set": bson.M{
			"last_login_at": acc.LastLoginAt,
			"login_token":   acc.Token,
		}})
	})
	if err != nil {
		log.Error("update token error:", err)
		return nil, err
	}

	log.Debugf("account id:%v new token:%v login time:%v",
		acc.UserId, acc.Token, acc.LastLoginAt.String())

	return &acc, nil
}
