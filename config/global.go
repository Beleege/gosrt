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
	PoolSize int    `default:"10"`
	UDP      struct {
		IP   string `default:"127.0.0.1"`
		Port int    `default:"9090"`
	}
	SRT struct {
		Latency struct {
			TX uint16 `default:"120"`
			RX uint16 `default:"20"`
		}
	}
	HLS struct {
		Server struct {
			Port int `default:"9091"`
		}
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

func GetHLSPort() int {
	return params.HLS.Server.Port
}

func GetPoolSize() int {
	return params.PoolSize
}

func GetTx() uint16 {
	return params.SRT.Latency.TX
}

func GetRx() uint16 {
	return params.SRT.Latency.RX
}
