package main

import (
	"common/models"
	"errors"
	"os"
	"regexp"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	db "lib/db/mongodb"
	"lib/proto/auth"
	"lib/proto/snowflake"
	sp "lib/services"
	"lib/utils"
)

const (
	SERVICE           = "[AUTH]"
	DefaultMongodbURL = "mongodb://172.17.0.1/account"
	EnvMongodb        = "MONGODB_URL"
	DBAccount         = "account"
)

var (
	ErrMethodNotSupported    = errors.New("method not supported")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrPasswordInvalid       = errors.New("password invalid")

	AuthFailResult = &auth.Auth_Result{OK: false, UserId: 0, Body: nil}

	uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

var (
	sfCli snowflake.SnowflakeServiceClient
	accDB *db.DialContext
)

type server struct {
}

func (s *server) init() {
	// 连接snowflake
	conn, _ := sp.GetService(sp.DefaultServicePath + "/snowflake")
	if conn == nil {
		log.Panic("cannot get snowflake service")
		os.Exit(-1)
	}
	sfCli = snowflake.NewSnowflakeServiceClient(conn)

	// 连接db
	mongodbURL := DefaultMongodbURL
	mkey, mval := sp.GetExtraService(sp.DefaultServicePath + "/mongo")
	if mkey == "" || mval == "" {
		if env := os.Getenv(EnvMongodb); env != "" {
			mongodbURL = env
		}
	} else {
		mongodbURL = "mongodb://" + mval
	}

	var err error
	accDB, err = db.Dial(mongodbURL, db.DefaultConcurrent)
	if err != nil {
		log.Panicf("mongodb: cannot connect to %v, err: %v", mongodbURL, err)
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
		var p = models.Account{}
		if err := msgpack.Unmarshal(cert.Proof, &p); err != nil {
			log.Error("Auth plain invalid, proof:", utils.Bytes2Str(cert.Proof))
			return AuthFailResult, nil
		}

		// 用户名密码验证
		acc, err := authOrReg(&p)
		if err != nil {
			return AuthFailResult, nil
		}

		return &auth.Auth_Result{
			OK:     true,
			UserId: p.UserId,
			Body:   utils.Str2Bytes(acc.Token)}, nil

	case auth.Auth_TOKEN:
	case auth.Auth_FACEBOOK:
	default:
		return nil, ErrMethodNotSupported
	}
	return nil, ErrMethodNotSupported
}
