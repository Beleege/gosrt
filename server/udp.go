package server

import (
	"fmt"
	"github.com/beleege/gosrt/core/selector"
	"net"

	"github.com/beleege/gosrt/config"
	"github.com/pkg/errors"
)

func SetupUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", config.GetUDPAddr())
	if err != nil {
		panic(errors.WithStack(err))
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(errors.WithStack(err))
	}
	defer func() {
		_ = conn.Close()
		fmt.Println("udp server shutdown")
	}()

	selector.Select(conn)
}
