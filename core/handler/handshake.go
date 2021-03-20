package handler

import (
	"net"

	"github.com/beleege/gosrt/util/log"
)

type HandShake struct {
	nextHandler SRTHandler
}

func NewHandShakeHandler() *HandShake {
	h := new(HandShake)
	return h
}

func (h *HandShake) hasNext() bool {
	return h.nextHandler != nil
}

func (h *HandShake) next(next SRTHandler) {
	h.nextHandler = next
}

func (h *HandShake) In(c net.PacketConn, b []byte) error {
	log.Info("receive handshake data: %s", string(b))
	return nil
}

func (h *HandShake) Out(c net.PacketConn, a net.Addr, b []byte) error {
	_, err := c.WriteTo(b, a)
	return err
}
