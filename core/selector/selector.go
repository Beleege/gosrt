package selector

import (
	"github.com/beleege/gosrt/core/handler"
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
				s = session.NewSRTSession(c, from, buf[:n])
				h.sessions[client] = s
			}

			handler.Queue <- s
		} else {
			//l.notifyReadError(errors.WithStack(err))
			return
		}
	}
}
