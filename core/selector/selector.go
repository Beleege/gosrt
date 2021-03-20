package selector

import (
	"net"

	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/util/log"
)

const (
	mtuLimit = 1500
)

type holder struct {
	sessions map[string]*session.SRTSession
}

func Select(c net.PacketConn) {
	h := new(holder)
	h.sessions = make(map[string]*session.SRTSession)

	buf := make([]byte, mtuLimit)

	for {
		if n, from, err := c.ReadFrom(buf); err == nil {
			client := from.String()
			log.Debug("conn build from: %s", client)

			s, ok := h.sessions[client]
			if !ok {
				s = session.NewSRTSession(c, from)
				h.sessions[client] = s
			}

			go work(s, buf[:n])
		} else {
			//l.notifyReadError(errors.WithStack(err))
			return
		}
	}
}

func work(s *session.SRTSession, b []byte) {
	if _, err := s.Read(b); err != nil {
		log.Error("read data fail: %s", err.Error())
	}
	if _, err := s.Write(b); err != nil {
		log.Error("write data fail: %s", err.Error())
	}
}
