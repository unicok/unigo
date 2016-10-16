package logger

import (
	"log"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
)

const (
	logLvEnv     = "LOG_LEVEL"
	defaultLogLv = logrus.InfoLevel
)

var (
	logLv logrus.Level
)

func init() {
	logLv = defaultLogLv
	if env := os.Getenv(logLvEnv); env != "" {
		v, err := strconv.ParseUint(env, 10, 8)
		if err == nil {
			logLv = logrus.Level(v)
		}
	}

	logrus.SetLevel(logLv)
	log.Println("Log Level:", logLv)
}
