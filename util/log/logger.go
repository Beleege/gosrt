package log

import (
	"github.com/beleege/gosrt/config"
	"os"

	logger "github.com/sirupsen/logrus"
)

var l *logger.Logger

func InitLog() {
	level, err := logger.ParseLevel(config.GetLogLevel())
	if err != nil {
		panic(err)
	}
	formatter := logger.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.999"}
	l = &logger.Logger{
		Out:       os.Stdout,
		Formatter: &formatter,
		Hooks:     make(logger.LevelHooks),
		Level:     level,
	}
}

func Info(s string, args ...interface{}) {
	l.Infof(s, args...)
}

func Debug(s string, args ...interface{}) {
	l.Debugf(s, args...)
}

func Error(s string, args ...interface{}) {
	l.Errorf(s, args...)
}
