package utils

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

// PrintPanicStack is print the call stack
func PrintPanicStack(extras ...interface{}) {
	if x := recover(); x != nil {
		log.Error(x)
		i := 0
		funcName, file, line, ok := runtime.Caller(i)
		for ok {
			log.WithFields(log.Fields{
				"frame": i,
				"func":  runtime.FuncForPC(funcName).Name(),
				"file":  file,
				"line":  line,
			}).Error("Panic stack")
			i++
			funcName, file, line, ok = runtime.Caller(i)
		}

		for k := range extras {
			log.WithFields(log.Fields{
				"exras": k,
				"data":  spew.Sdump(extras[k]),
			}).Error("extras dump")
		}
	}
}
