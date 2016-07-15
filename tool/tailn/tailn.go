package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/pquerna/ffjson/ffjson"
	log "github.com/unicok/unigo/lib/nsq-logger"
	"gopkg.in/urfave/cli.v2"
)

const (
	layout = "1985/09/02 00:00:00"
)

var (
	logTemplate = make(map[byte]string)
	signCh      = make(chan os.Signal, 1)
)

func init() {
	logTemplate[log.FINEST] = "\033[1;34m%v [FINEST] %v %v %v %v %v \033[0m"
	logTemplate[log.FINE] = "\033[0;34m%v [FINE] %v %v %v %v %v \033[0m"
	logTemplate[log.DEBUG] = "\033[1;32m%v [DEBUG] %v %v %v %v %v \033[0m"
	logTemplate[log.TRACE] = "\033[0;37m%v [TRACE] %v %v %v %v %v \033[0m"
	logTemplate[log.WARN] = "\033[0;33m%v [WARNING] %v %v %v %v %v \033[0m"
	logTemplate[log.INFO] = "\033[0;32m%v [INFO] %v %v %v %v %v  \033[0m"
	logTemplate[log.ERROR] = "\033[0;31m%v [ERROR] %v %v %v %v %v \033[0m"
	logTemplate[log.CRITICAL] = "\033[7;31m%v [CRITICAL] %v %v %v %v %v \033[0m"
}

var (
	tailFlag = []cli.Flag{
		&cli.StringFlag{
			Name:    "topic",
			Aliases: []string{"t"},
			Value:   "LOG",
			Usage:   "NSQ topic, default is LOG",
		},
		&cli.StringFlag{
			Name:    "channel",
			Aliases: []string{"c"},
			Value:   "tailn",
			Usage:   "NSQ channel, default is tailn",
		},
		&cli.StringFlag{
			Name:    "number",
			Aliases: []string{"n"},
			Value:   "0",
			Usage:   "Line to show, default no limit",
		},
		&cli.StringFlag{
			Name:    "nsqd-tcp-address",
			Aliases: []string{"a"},
			Value:   "localhost:4150",
			Usage:   "nsqd TCP address",
		},
		&cli.StringFlag{
			Name:    "lookup-http-address",
			Aliases: []string{"l"},
			Value:   "localhost:4161",
			Usage:   "lookupd HTTP address",
		},
		&cli.StringFlag{
			Name:    "timeout",
			Aliases: []string{"o"},
			Value:   "5",
			Usage:   "Dial timeout, default 5s",
		},
		&cli.StringFlag{
			Name:  "type",
			Value: "NSQLOG",
			Usage: "Tail type , default NSQLOG (others print json )",
		},
		&cli.StringFlag{
			Name:  "log",
			Value: "false",
			Usage: "whether open inner log",
		},
		&cli.StringFlag{
			Name:    "tofile",
			Aliases: []string{"f"},
			Value:   "",
			Usage:   "output to file, defualt is write into current dir",
		},
	}
)

// TailHandler is a tail handler struct
type TailHandler struct {
	totalMessages int
	messagesShown int
	writer        io.Writer
	printMessage  func(io.Writer, *nsq.Message) error
}

func nsqLog(w io.Writer, m *nsq.Message) error {
	info := &log.LogFormat{}
	err := ffjson.Unmarshal(m.Body, &info)
	if err != nil {
		fmt.Printf("err %v\n", err)
		return nil
	}
	_, err = fmt.Fprintln(w, fmt.Sprintf(logTemplate[info.Level], info.Time.Format(layout), info.Prefix, info.Host, info.Msg, info.Caller, info.LineNo))
	if err != nil {
		return err
	}
	return nil
}

func stdLog(w io.Writer, m *nsq.Message) error {
	_, err := fmt.Fprintln(w, m.Body)
	if err != nil {
		return err
	}
	return nil
}

func fileLog(w io.Writer, m *nsq.Message) error {
	info := &log.LogFormat{}
	err := ffjson.Unmarshal(m.Body, &info)
	if err != nil {
		fmt.Printf("err %v\n", err)
		return nil
	}
	line := fmt.Sprintf("%v [%v] %v %v %v %v %v", info.Time.Format(layout), info.Level, info.Prefix, info.Host, info.Msg, info.Caller, info.LineNo)
	_, err = fmt.Fprintln(w, line)
	if err != nil {
		return err
	}
	return nil
}

func doAction(c *cli.Context) error {
	// TODO CheckFlag
	checkFlag(c)
	_cfg := nsq.NewConfig()
	// dail timeout
	_cfg.DialTimeout = time.Duration(c.Int("timeout")) * time.Second
	_cfg.UserAgent = fmt.Sprintf("go-nsq version:%v", nsq.VERSION)
	_cfg.MaxInFlight = 128
	_consumer, err := nsq.NewConsumer(c.String("topic"), c.String("channel"), _cfg)
	if err != nil {
		fmt.Printf("error %v\n", err)
		os.Exit(0)
	}

	_f := nsqLog
	var _w io.Writer = os.Stdout
	if c.String("type") != "NSGLOG" {
		_f = stdLog
	}

	if c.String("tofile") != "" {
		_f = fileLog
		fl, err := os.OpenFile(c.String("tofile"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("error %v\n", err)
			os.Exit(0)
		}
		_w = fl
	}

	// disable inner log
	if !c.Bool("log") {
		_consumer.SetLogger(nil, 0)
	}

	_consumer.AddHandler(&TailHandler{c.Int("number"), 0, _w, _f})
	err = _consumer.ConnectToNSQDs(strings.Split(c.String("nsqd-tcp-addres"), ","))
	if err != nil {
		fmt.Printf("error %v\n", err)
		os.Exit(0)
	}
	err = _consumer.ConnectToNSQLookupds(strings.Split(c.String("lookupd-http-address"), ","))
	if err != nil {
		fmt.Printf("error %v\n", err)
		os.Exit(0)
	}
	for {
		select {
		case <-_consumer.StopChan:
			return nil
		case <-signCh:
			_consumer.Stop()
		}
	}
}

func main() {
	app := &cli.App{
		Name:    "tailn",
		Usage:   "Tail log from nsq!",
		Version: "0.0.1",
		Flags:   tailFlag,
		Action:  doAction,
	}

	signal.Notify(signCh, syscall.SIGINT, syscall.SIGTERM)
	app.Run(os.Args)
}

func checkFlag(c *cli.Context) {
	if c.String("channel") == "" || c.String("topic") == "" {
		cli.ShowAppHelp(c)
		os.Exit(0)
	}
}

// HandleMessage is the message processing interface for Consumer
func (th *TailHandler) HandleMessage(m *nsq.Message) error {
	th.messagesShown++
	if err := th.printMessage(th.writer, m); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}
	if th.totalMessages > 0 && th.totalMessages < th.messagesShown {
		os.Exit(0)
	}
	return nil
}
