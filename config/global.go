package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"
)

var params = struct {
	LogLevel string `default:"info" env:"Loglevel"`
	LogFile  string `default:"/tmp/debug.log" env:"LogFile"`
	UDP      struct {
		IP   string `default:"127.0.0.1"`
		Port int    `default:"9090"`
	}
}{}

func InitConfig() {
	pwd, _ := os.Getwd()
	file := "property.yaml"
	path := filepath.Join(pwd, file)

	if err := configor.Load(&params, path); err != nil {
		panic(err)
	}
}

func GetLogLevel() string {
	return params.LogLevel
}

func GetUDPAddr() string {
	return fmt.Sprintf("%s:%d", params.UDP.IP, params.UDP.Port)
}
