package log

import (
	"os"

	"github.com/beleege/gosrt/config"
	logger "github.com/sirupsen/logrus"
)

var l *logger.Logger

func InitLog() {
	level, err := logger.ParseLevel(config.GetLogLevel())
	if err != nil {
		panic(err)
	}
	formatter := logger.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.999999"}
	l = &logger.Logger{
		Out:       os.Stdout,
		Formatter: &formatter,
		Hooks:     make(logger.LevelHooks),
		Level:     level,
	}
}

func Infof(s string, args ...interface{}) {
	l.Infof(s, args...)
}

func Debugf(s string, args ...interface{}) {
	l.Debugf(s, args...)
}

func Errorf(s string, args ...interface{}) {
	l.Errorf(s, args...)
}
