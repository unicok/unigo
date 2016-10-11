package main

import (
	"common/models"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"

	"gopkg.in/mgo.v2"

	log "github.com/Sirupsen/logrus"

	"golang.org/x/net/context"

	db "lib/db/mongodb"
	"lib/proto/auth"
	"lib/proto/snowflake"
	sp "lib/services"
	"lib/utils"
)

const (
	SERVICE           = "[CHAT]"
	DefaultMongodbURL = "mongodb://172.17.0.1/account"
	EnvMongodb        = "MONGODB_URL"
	DBAccount         = "account"
)

var (
	ErrorMethodNotSupported = errors.New("method not supported")

	AuthFailResult = &auth.Auth_Result{OK: false, UserId: 0, Body: nil}

	uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

type server struct {
	sfClient snowflake.SnowflakeServiceClient
	db       *db.DialContext
}

func (s *server) init() {
	// 连接snowflake
	conn, _ := sp.GetService(sp.DefaultServicePath + "/snowflake")
	if conn == nil {
		log.Panic("cannot get snowflake service")
		os.Exit(-1)
	}
	s.sfClient = snowflake.NewSnowflakeServiceClient(conn)

	// 连接db
	mongodbURL := DefaultMongodbURL
	if env := os.Getenv(EnvMongodb); env != "" {
		mongodbURL = env
	}

	var err error
	s.db, err = db.Dial(mongodbURL, db.DefaultConcurrent)
	if err != nil {
		log.Panic("mongodb: cannot connect to", mongodbURL, err)
		os.Exit(-1)
	}
}

func (s *server) Auth(ctx context.Context, cert *auth.Auth_Certificate) (*auth.Auth_Result, error) {
	switch cert.Type {
	case auth.Auth_UUID:
		if uuidRegexp.MatchString(strings.ToLower(string(cert.Proof))) {
			return AuthFailResult, nil
		}
		return &auth.Auth_Result{OK: true, UserId: 0, Body: nil}, nil
	case auth.Auth_PLAIN:
		var p struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(cert.Proof, &p); err != nil {
			log.Error("Auth plain invalid, proof:", utils.Bytes2Str(cert.Proof))
			return AuthFailResult, nil
		}

		// 登录验证步骤
		// 1.查询数据库,验证用户名密码 验证成功则直接返回
		// 2.如果不成功,获取自增ID
		// 3.新建用户数据 并返回

		// 验证用户
		var err error
		var account models.Account
		err = s.db.DBAction(DBAccount, "account", func(c *mgo.Collection) error {
			return c.Find(p.Username).One(&account)
		})
		if err != nil {
			return AuthFailResult, nil
		}

		if account.Pass == p.Password {
			return &auth.Auth_Result{OK: true, UserId: account.UserId, Body: nil}, nil
		}

		// 获取ID
		res, err := s.sfClient.Next(context.Background(), &snowflake.Snowflake_Key{Name: "userid"})
		if err != nil {
			return AuthFailResult, nil
		}

		log.Info("new userid:", res.Value)

		// 创建用户

	case auth.Auth_TOKEN:
	case auth.Auth_FACEBOOK:
	default:
		return nil, ErrorMethodNotSupported
	}
	return nil, ErrorMethodNotSupported
}
