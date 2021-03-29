package selector

import (
	"net"
	"sync"

	"github.com/beleege/gosrt/core/handler"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/util/log"
)

const (
	_mtuLimit = 1500
)

type holder struct {
	sessions *sync.Map
}

func Select(c net.PacketConn) {
	h := new(holder)
	h.sessions = &sync.Map{}

	buf := make([]byte, _mtuLimit)

	defer func() {
		defer close(handler.Queue)
	}()

	for {
		if n, from, err := c.ReadFrom(buf); err == nil {
			client := from.String()
			log.Debugf("conn build from: %s", client)

			var s *session.SRTSession
			v, ok := h.sessions.Load(client)
			if !ok {
				s = session.NewSRTSession(c, from, buf[:n])
				h.sessions.Store(client, s)
			} else {
				s = v.(*session.SRTSession)
				s.Data = buf[:n]
			}

			handler.Queue <- s
		} else {
			//l.notifyReadError(errors.WithStack(err))
			return
		}
	}
}
