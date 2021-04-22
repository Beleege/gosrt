package selector

import (
	"encoding/binary"
	"net"

	"github.com/beleege/gosrt/core/handler"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/util/log"
)

const (
	_mtuLimit = 1400
)

type holder struct {
	sessions map[string]*session.SRTSession
}

func Select(conn net.PacketConn) {
	h := new(holder)
	h.sessions = make(map[string]*session.SRTSession)

	defer func() {
		close(handler.Queue)
	}()

	for {
		buf := make([]byte, _mtuLimit)
		if n, from, err := conn.ReadFrom(buf); err == nil {
			client := from.String()
			log.Debugf(">>>>>>> package seqNo: %d", binary.BigEndian.Uint32(buf[:4]))

			s := h.sessions[client]
			if s == nil {
				log.Infof("########## create session for %s", client)
				s = session.NewSRTSession(conn, from)
				h.sessions[client] = s
			}

			handler.Queue <- handler.NewBox(s, buf[:n])
		} else {
			//l.notifyReadError(errors.WithStack(err))
			return
		}
	}
}
