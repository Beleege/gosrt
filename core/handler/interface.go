package handler

import (
	"net"
)

type SRTHandler interface {
	hasNext() bool
	next(h SRTHandler)
	In(s net.PacketConn, b []byte) error
	Out(s net.PacketConn, a net.Addr, b []byte) error
}

func Warp(handlers ...SRTHandler) SRTHandler {
	head := handlers[len(handlers)-1]
	next := head
	for i := len(handlers) - 2; i >= 0; i-- {
		next.next(handlers[i])
		next = handlers[i]
	}
	return head
}
