package logger

import (
	"log"
	"os"
)

// Logger接口
type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

var (
	ERROR Logger = log.New(os.Stderr, "mqtt-broker[error]", log.Lshortfile|log.Ldate|log.Ltime)
	WARN  Logger = log.New(os.Stdout, "mqtt-broker[warn]", log.Lshortfile|log.Ldate|log.Ltime)
	INFO  Logger = log.New(os.Stdout, "mqtt-broker[info]", log.Lshortfile|log.Ldate|log.Ltime)
	DEBUG Logger = log.New(os.Stdout, "mqtt-broker[debug]", log.Lshortfile|log.Ldate|log.Ltime)
)
