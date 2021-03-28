package handler

import (
	"encoding/binary"
	"time"

	"github.com/beleege/gosrt/config"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/beleege/gosrt/util/math"
	"github.com/pkg/errors"
)

type handshake struct {
	nextHandler srtHandler
}

func NewHandshake() *handshake {
	h := new(handshake)
	return h
}

func (h *handshake) hasNext() bool {
	return h.nextHandler != nil
}

func (h *handshake) next(next srtHandler) {
	h.nextHandler = next
}

func (h *handshake) execute(s *session.SRTSession) error {
	if s.CP != nil && s.CP.CType == srt.CTHandShake && s.Status != session.SConnect {
		if s.Status == session.SOpen {
			return responseAndSetCookie(s)
		} else if s.Status == session.SRepeat {
			return establishConnection(s)
		} else {
			log.Infof("illegal handshake status: %d", s.Status)
			s.Status = session.SShutdown
			return errors.Errorf("handshake fail")
		}
	} else if h.hasNext() {
		return h.nextHandler.execute(s)
	}
	return errors.New("no handler after handshake")
}

func responseAndSetCookie(s *session.SRTSession) error {
	ipv4, err := s.GetPeerIPv4()
	if err != nil {
		return nil
	}
	binary.BigEndian.PutUint32(s.Data[8:12], uint32(time.Now().UnixNano()-s.OpenTime.UnixNano()))
	binary.BigEndian.PutUint32(s.Data[12:16], s.ThatSID)
	binary.BigEndian.PutUint32(s.Data[16:20], uint32(5))
	binary.BigEndian.PutUint16(s.Data[20:22], uint16(0))
	binary.BigEndian.PutUint16(s.Data[22:24], uint16(srt.HSv5Magic))
	binary.BigEndian.PutUint32(s.Data[24:28], s.SendNo)
	//binary.BigEndian.PutUint32(s.Data[28:32], s.MTU)
	//binary.BigEndian.PutUint32(s.Data[32:36], s.MFW)
	//binary.BigEndian.PutUint32(s.Data[36:40], s.HT)
	binary.BigEndian.PutUint32(s.Data[40:44], s.ThatSID)
	buildCookie(s, ipv4)
	s.Data[48] = ipv4[3]
	s.Data[49] = ipv4[2]
	s.Data[50] = ipv4[1]
	s.Data[51] = ipv4[0]

	if _, err = s.Write(s.Data[:64]); err != nil {
		return err
	}
	s.Status = session.SSetCookie
	return nil
}

func buildCookie(s *session.SRTSession, ipv4 *[4]byte) {
	now := s.OpenTime.Minute()
	s.Data[47] = ipv4[2]
	s.Data[46] = ipv4[3]
	s.Data[46] = byte(now >> 8)
	s.Data[44] = byte(now)
	s.Cookie = binary.BigEndian.Uint32(s.Data[44:48])
	log.Infof("response [%s] with cookie[%d]", s.GetPeer(), s.Cookie)
}

func establishConnection(s *session.SRTSession) error {
	ipv4, err := s.GetPeerIPv4()
	if err != nil {
		return nil
	}
	binary.BigEndian.PutUint32(s.Data[8:12], uint32(0))
	binary.BigEndian.PutUint32(s.Data[12:16], s.ThatSID)
	binary.BigEndian.PutUint16(s.Data[22:24], uint16(1))
	binary.BigEndian.PutUint32(s.Data[40:44], s.ThisSID)
	s.Data[48] = ipv4[3]
	s.Data[49] = ipv4[2]
	s.Data[50] = ipv4[1]
	s.Data[51] = ipv4[0]

	binary.BigEndian.PutUint16(s.Data[64:66], uint16(srt.HSExtTypeHSRsp))
	binary.BigEndian.PutUint16(s.Data[76:78], math.MaxUInt16(s.TSBPD.TxDelay, config.GetRx()))
	binary.BigEndian.PutUint16(s.Data[78:80], math.MaxUInt16(s.TSBPD.RxDelay, config.GetTx()))

	if _, err = s.Write(s.Data[:80]); err != nil {
		return err
	}
	s.Status = session.SConnect
	return nil
}
