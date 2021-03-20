package session

import (
	"net"

	"github.com/beleege/gosrt/core/handler"
)

type SRTSession struct {
	conn    net.PacketConn
	peer    net.Addr
	handler handler.SRTHandler
}

func (s *SRTSession) Read(b []byte) (n int, err error) {
	return 0, s.handler.In(s.conn, b)
}

func (s *SRTSession) Write(b []byte) (n int, err error) {
	return 0, s.handler.Out(s.conn, s.peer, b)
}

func NewSRTSession(c net.PacketConn, a net.Addr) *SRTSession {
	s := new(SRTSession)
	s.conn = c
	s.peer = a
	s.handler = handler.Warp(selectHandlers()...)
	return s
}

func selectHandlers() []handler.SRTHandler {
	list := make([]handler.SRTHandler, 0, 1)
	list = append(list, handler.NewHandShakeHandler())
	return list
}
