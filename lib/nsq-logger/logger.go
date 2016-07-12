package nsqlogger

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	nsqdEnv       = "NSQD_HOST"
	defaultPubURL = "http://172.17.42.1:4151/mpub?topic=LOG&binary=true"
	mine          = "application/octet-strem"
)

const (
	//FINEST level
	FINEST byte = iota
	//FINE level
	FINE
	//DEBUG level
	DEBUG
	//TRACE level
	TRACE
	//INFO level
	INFO
	//WARN level
	WARN
	//ERROR level
	ERROR
	//CRITICAL level
	CRITICAL
)

//LogFormat is store log content
type LogFormat struct {
	Prefix string
	Time   time.Time
	Host   string
	Level  byte
	Msg    string
	Caller string
	LineNo int
}

var (
	pubAddr string
	prefix  string
	level   byte
	ch      chan []byte
	flushCh chan chan struct{}
)

func init() {
	pubAddr = defaultPubURL
	if env := os.Getenv(nsqdEnv); env != "" {
		pubAddr = env + "/mpub?topic=LOG&binary=true"
	}
	ch = make(chan []byte, 4096)
	flushCh = make(chan chan struct{}, 1)
	go publishTask()
}

func publishTask() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	size := make([]byte, 4)

	flush := func() {
		n := len(ch)
		if n == 0 {
			return
		}

		// [ 4-byte num messages ]
		// [ 4-byte message #1 size ] [ N-byte binary data ]
		// ... (repeated <num_messages> times)
		buf := new(bytes.Buffer)
		binary.BigEndian.PutUint32(size, uint32(n))
		buf.Write(size)
		for i := 0; i < n; i++ {
			bts := <-ch
			binary.BigEndian.PutUint32(size, uint32(len(bts)))
			buf.Write(size)
			buf.Write(bts)
		}

		// http post data
		resp, err := http.Post(pubAddr, mine, buf)
		if err != nil {
			log.Println(err)
			return
		}
		if _, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Println(err)
		}
		resp.Body.Close()
	}

	for {
		select {
		case <-ticker.C:
			flush()
		case w := <-flushCh:
			flush()
			w <- struct{}{}
		}
	}
}

// publish to nsqd (localhost nsqd is suggested!)
func publish(msg LogFormat) {
	// fill in the common fields
	hostname, _ := os.Hostname()
	msg.Host = hostname
	msg.Time = time.Now()
	msg.Prefix = prefix

	// Determine caller func
	if pc, _, lineno, ok := runtime.Caller(2); ok {
		msg.Caller = runtime.FuncForPC(pc).Name()
		msg.LineNo = lineno
	}

	// pack message
	if bts, err := ffjson.Marshal(msg); err == nil {
		fmt.Println(msg)
		ch <- bts
	} else {
		log.Println(err, msg)
		return
	}
}

// Flush remaining logs
func Flush() {
	w := make(chan struct{}, 1)
	flushCh <- w
	<-w
}

// SetPrefix to set prefix of log
func SetPrefix(pfix string) {
	prefix = pfix
}

// SetLogLevel to set log level
func SetLogLevel(lv byte) {
	level = lv
}

// Finest is wrapper for finest log level
func Finest(v ...interface{}) {
	if level <= FINEST {
		msg := LogFormat{Level: FINEST, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Finestf is wrapper for finest log level
func Finestf(format string, v ...interface{}) {
	if level <= FINEST {
		msg := LogFormat{Level: FINEST, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Fine is wrapper for finest log level
func Fine(v ...interface{}) {
	if level <= FINE {
		msg := LogFormat{Level: FINE, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Finef is wrapper for finest log level
func Finef(format string, v ...interface{}) {
	if level <= FINE {
		msg := LogFormat{Level: FINE, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Debug is wrapper for debug log level
func Debug(v ...interface{}) {
	if level <= DEBUG {
		msg := LogFormat{Level: DEBUG, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Debugf is wrapper for debug log level
func Debugf(format string, v ...interface{}) {
	if level <= DEBUG {
		msg := LogFormat{Level: DEBUG, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Trace is wrapper for trace log level
func Trace(v ...interface{}) {
	if level <= TRACE {
		msg := LogFormat{Level: TRACE, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Tracef is wrapper for trace log level
func Tracef(format string, v ...interface{}) {
	if level <= TRACE {
		msg := LogFormat{Level: TRACE, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Info is wrapper for info log level
func Info(v ...interface{}) {
	if level <= INFO {
		msg := LogFormat{Level: INFO, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Infof is wrapper for info log level
func Infof(format string, v ...interface{}) {
	if level <= INFO {
		msg := LogFormat{Level: INFO, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Warn is wrapper for warn log level
func Warn(v ...interface{}) {
	if level <= WARN {
		msg := LogFormat{Level: WARN, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Warnf is wrapper for warn log level
func Warnf(format string, v ...interface{}) {
	if level <= WARN {
		msg := LogFormat{Level: WARN, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Error is wrapper for error log level
func Error(v ...interface{}) {
	if level <= ERROR {
		msg := LogFormat{Level: ERROR, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Errorf is wrapper for error log level
func Errorf(format string, v ...interface{}) {
	if level <= ERROR {
		msg := LogFormat{Level: ERROR, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}

// Critical is wrapper for critical log level
func Critical(v ...interface{}) {
	if level <= CRITICAL {
		msg := LogFormat{Level: CRITICAL, Msg: fmt.Sprint(v...)}
		publish(msg)
	}
}

// Criticalf is wrapper for critical log level
func Criticalf(format string, v ...interface{}) {
	if level <= CRITICAL {
		msg := LogFormat{Level: CRITICAL, Msg: fmt.Sprintf(format, v...)}
		publish(msg)
	}
}
