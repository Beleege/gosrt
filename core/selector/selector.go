package selector

import (
	"net"
	"sync"

	"github.com/beleege/gosrt/core/handler"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/util/log"
)

const (
	_mtuLimit = 1400
)

var (
	_sessions map[string]*session.SRTSession
	_pool     *sync.Pool
)

func Select(conn net.PacketConn) {
	_sessions = make(map[string]*session.SRTSession)
	//_pool = &sync.Pool{New: newBuf}

	defer func() {
		close(handler.Queue)
	}()

	for {
		buf := make([]byte, _mtuLimit)
		if n, from, err := conn.ReadFrom(buf); err == nil {
			client := from.String()

			s := _sessions[client]
			if s == nil {
				log.Infof("########## create session for %s", client)
				s = session.NewSRTSession(conn, from)
				_sessions[client] = s
			}

			handler.Queue <- handler.NewBox(s, buf[:n])
		} else {
			//l.notifyReadError(errors.WithStack(err))
			return
		}
	}
}

func GetAllSession() (list []*session.SRTSession) {
	for k, _ := range _sessions {
		list = append(list, _sessions[k])
	}
	return
}

func Recycle(d []byte) {
	_pool.Put(d)
}

func newBuf() interface{} {
	return make([]byte, _mtuLimit)
}
