package main

import (
	"errors"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"golang.org/x/net/context"

	"lib/db/mgo"
	pb "lib/proto/auth"
	sf "lib/proto/snowflake"
	sp "lib/services"
)

const (
	SERVICE           = "[CHAT]"
	DefaultMongodbURL = "mongodb://172.17.0.1/account"
	EnvMongodb        = "MONGODB_URL"
	DBAccount         = "account"
)

var (
	ErrorMethodNotSupported = errors.New("method not supported")

	uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

type server struct {
	sfClient sf.SnowflakeServiceClient
	db       *mgo.DialContext
}

func (s *server) init() {
	// 连接snowflake
	conn, _ := sp.GetService(sp.DefaultServicePath + "/snowflake")
	if conn == nil {
		log.Panic("cannot get snowflake service")
		os.Exit(-1)
	}
	s.sfClient = sf.NewSnowflakeServiceClient(conn)

	// 连接db
	mongodbURL := DefaultMongodbURL
	if env := os.Getenv(EnvMongodb); env != "" {
		mongodbURL = env
	}

	var err error
	s.db, err = mgo.Dial(mongodbURL, mgo.DefaultConcurrent)
	if err != nil {
		log.Panic("mongodb: cannot connect to", mongodbURL, err)
		os.Exit(-1)
	}
}

func (s *server) Auth(ctx context.Context, cert *pb.Auth_Certificate) (*pb.Auth_Result, error) {
	switch cert.Type {
	case pb.Auth_UUID:
		if uuidRegexp.MatchString(strings.ToLower(string(cert.Proof))) {
			return &pb.Auth_Result{OK: true, UserId: 0, Body: nil}, nil
		}
		return &pb.Auth_Result{OK: true, UserId: 0, Body: nil}, nil
	case pb.Auth_PLAIN:
		// 登录验证步骤
		// 1.查询数据库,验证用户名密码 验证成功则直接返回
		// 2.如果不成功,获取自增ID
		// 3.新建用户数据 并返回

		// 验证用户
		var err error
		s.db.Query(func(sess *mgo.Session) error {
			// sess.DB(DBAccount).C("user").Find("")
			return nil
		})
		// 获取ID
		res, err := s.sfClient.Next(context.Background(), &sf.Snowflake_Key{Name: "userid"})
		if err != nil {
			return &pb.Auth_Result{OK: false, UserId: 0, Body: nil}, nil
		}

		log.Info("new userid:", res.Value)
		// 创建用户

	case pb.Auth_TOKEN:
	case pb.Auth_FACEBOOK:
	default:
		return nil, ErrorMethodNotSupported
	}
	return nil, ErrorMethodNotSupported
}
