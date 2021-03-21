package handler

import (
	"encoding/binary"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
)

type greeter struct {
	nextHandler srtHandler
}

func NewGreeter() *greeter {
	h := new(greeter)
	return h
}

func (h *greeter) hasNext() bool {
	return h.nextHandler != nil
}

func (h *greeter) next(next srtHandler) {
	h.nextHandler = next
}

func (h *greeter) execute(s *session.SRTSession) error {
	if s.CP != nil && s.CP.CType == srt.CTHandShake {
		if s.Status == session.SOpen {
			if err := responseAndSetCookie(s); err != nil {
				return err
			}
		} else if s.Status == session.SRepeat {

		}
	} else if h.hasNext() {
		return h.nextHandler.execute(s)
	}
	return nil
}

func responseAndSetCookie(s *session.SRTSession) error {
	binary.BigEndian.PutUint32(s.Data[8:12], uint32(0x01010101))
	binary.BigEndian.PutUint32(s.Data[12:16], s.ThatSID)
	binary.BigEndian.PutUint32(s.Data[16:20], uint32(5))
	binary.BigEndian.PutUint16(s.Data[20:22], uint16(0))
	binary.BigEndian.PutUint16(s.Data[22:24], uint16(srt.HSv5Magic))
	binary.BigEndian.PutUint32(s.Data[28:32], uint32(0xCAFEBABE))
	ipv4, err := s.GetPeerIPv4()
	if err != nil {
		return nil
	}
	s.Data[32] = ipv4[3]
	s.Data[33] = ipv4[2]
	s.Data[34] = ipv4[1]
	s.Data[35] = ipv4[0]
	binary.BigEndian.PutUint32(s.Data[36:], uint32(0))

	if _, err = s.Write(s.Data); err != nil {
		return err
	}
	s.Status = session.SSetCookie
	return nil
}
