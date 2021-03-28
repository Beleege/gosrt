package server

import (
	"net"

	"github.com/beleege/gosrt/config"
	"github.com/beleege/gosrt/core/handler"
	"github.com/beleege/gosrt/core/selector"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

func SetupUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", config.GetUDPAddr())
	if err != nil {
		panic(errors.WithStack(err))
	}
	log.Infof("udp server start at %s", config.GetUDPAddr())

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(errors.WithStack(err))
	}
	defer func() {
		_ = conn.Close()
		log.Infof("udp server shutdown")
	}()

	go handler.Task()

	selector.Select(conn)
}
