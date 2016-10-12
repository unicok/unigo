package handler

import (
	"common/models"
	"crypto/rc4"
	"fmt"
	"io"
	"lib/crypto/dh"
	"lib/packet"
	"math/big"

	"gopkg.in/vmihailenco/msgpack.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	. "agent/types"
	"lib/proto/auth"
	pb "lib/proto/game"
	sp "lib/services"

	log "github.com/Sirupsen/logrus"
)

const (
	Salt        = "DH"
	DefaultGSID = "game1"
)

func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code[E_heart_beat_ack], tbl, nil)
}

// 密钥交换
// 加密建立方式: DH+RC4
// 注意:完整的加密过程包括 RSA+DH+RC4
// 1. RSA用于鉴定服务器的真伪(这步省略)
// 2. DH用于在不安全的信道上协商安全的KEY
// 3. RC4用于流加密
func P_get_seed_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_seed_info(reader)
	// KEY1
	X1, E1 := dh.DHExchange()
	KEY1 := dh.DHKey(X1, big.NewInt(int64(tbl.F_client_send_seed)))

	// KEY2
	X2, E2 := dh.DHExchange()
	KEY2 := dh.DHKey(X2, big.NewInt(int64(tbl.F_client_receive_seed)))

	ret := S_seed_info{int32(E1.Int64()), int32(E2.Int64())}
	// 服务器加密种子是客户端解密种子
	encoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", Salt, KEY2)))
	if err != nil {
		log.Error(err)
		return nil
	}
	decoder, err := rc4.NewCipher([]byte(fmt.Sprintf("%v%v", Salt, KEY1)))
	if err != nil {
		log.Error(err)
		return nil
	}
	sess.Encoder = encoder
	sess.Decoder = decoder
	sess.Flag |= SessKeyDone
	return packet.Pack(Code["get_seed_ack"], ret, nil)
}

// 玩家登陆过程
func P_user_login_req(sess *Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_user_login_info(reader)

	var acc = models.Account{
		UserName:   tbl.F_user_name,
		Password:   tbl.F_password_md5,
		OpenUUID:   tbl.F_open_udid,
		Lang:       tbl.F_user_lang,
		DeviceName: tbl.F_device_name,
		DeviceId:   tbl.F_device_id,
		DeviceType: tbl.F_device_id_type,
	}

	// TODO: 登陆鉴权
	// 简单鉴权可以在agent直接完成，通常公司都存在一个用户中心服务器用于鉴权
	authConn, serviceID := sp.GetService(sp.DefaultServicePath + "/auth")
	if authConn == nil {
		log.Error("cannot get auth service:", serviceID)
		return nil
	}
	authCli := auth.NewAuthServiceClient(authConn)
	proof, err := msgpack.Marshal(&acc)
	if err != nil {
		log.Errorf("msgpack marshal err:%v", err)
		return nil
	}

	authRes, err := authCli.Auth(context.Background(), &auth.Auth_Certificate{Type: 1, Proof: proof})
	if err != nil {
		log.Error(err)
		return nil
	}
	sess.UserID = authRes.UserId

	// TODO: 选择GAME服务器
	// 选服策略依据业务进行，比如小服可以固定选取某台，大服可以采用HASH或一致性HASH
	sess.GSID = DefaultGSID

	// 连接到已选定GAME服务器
	gsConn := sp.GetServieWithID(sp.DefaultServicePath+"/game", sess.GSID)
	if gsConn == nil {
		log.Error("cannot get game service:", sess.GSID)
		return nil
	}
	gsCli := pb.NewGameServiceClient(gsConn)

	// 开启到游戏服的流
	md := metadata.New(map[string]string{
		"userid": fmt.Sprint(sess.UserID),
	})
	ctx := metadata.NewContext(context.Background(), md)
	stream, err := gsCli.Stream(ctx)
	if err != nil {
		log.Error(err)
		return nil
	}
	sess.Stream = stream

	// 读取GAME返回消息的goroutine
	fetcher := func(sess *Session) {
		for {
			in, err := sess.Stream.Recv()
			if err == io.EOF {
				log.Debug(err)
				return
			} else if err != nil {
				log.Error(err)
				return
			}

			select {
			case sess.MQ <- *in:
			case <-sess.Die:
			}
		}
	}
	go fetcher(sess)
	return packet.Pack(Code[E_user_login_succeed_ack],
		S_user_snapshot{F_uid: sess.UserID},
		nil)
}
