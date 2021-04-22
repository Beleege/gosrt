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

func (h *handshake) execute(box *Box) error {
	if box.s.CP != nil && box.s.CP.CType == srt.CTHandShake && box.s.Status != session.SConnect {
		if box.s.Status == session.SOpen {
			return responseAndSetCookie(box)
		} else if box.s.Status == session.SRepeat {
			return establishConnection(box)
		} else {
			log.Infof("illegal handshake status: %d", box.s.Status)
			box.s.Status = session.SShutdown
			return errors.Errorf("handshake fail")
		}
	} else if h.hasNext() {
		return h.nextHandler.execute(box)
	}
	return errors.New("no handler after handshake")
}

func responseAndSetCookie(box *Box) error {
	ipv4, err := box.s.GetPeerIPv4()
	if err != nil {
		return nil
	}
	binary.BigEndian.PutUint32(box.b[8:12], uint32(time.Now().UnixNano()-box.s.OpenTime.UnixNano()))
	binary.BigEndian.PutUint32(box.b[12:16], box.s.ThatSID)
	binary.BigEndian.PutUint32(box.b[16:20], uint32(5))
	binary.BigEndian.PutUint16(box.b[20:22], uint16(0))
	binary.BigEndian.PutUint16(box.b[22:24], uint16(srt.HSv5Magic))
	binary.BigEndian.PutUint32(box.b[24:28], box.s.SendNo)
	//binary.BigEndian.PutUint32(s.Data[28:32], s.MTU)
	//binary.BigEndian.PutUint32(s.Data[32:36], s.MFW)
	//binary.BigEndian.PutUint32(s.Data[36:40], s.HT)
	binary.BigEndian.PutUint32(box.b[40:44], box.s.ThatSID)
	buildCookie(box, ipv4)
	box.b[48] = ipv4[3]
	box.b[49] = ipv4[2]
	box.b[50] = ipv4[1]
	box.b[51] = ipv4[0]

	if _, err = box.s.Write(box.b[:64]); err != nil {
		return err
	}
	box.s.Status = session.SSetCookie
	return nil
}

func buildCookie(box *Box, ipv4 *[4]byte) {
	now := box.s.OpenTime.Minute()
	box.b[47] = ipv4[2]
	box.b[46] = ipv4[3]
	box.b[46] = byte(now >> 8)
	box.b[44] = byte(now)
	box.s.Cookie = binary.BigEndian.Uint32(box.b[44:48])
	log.Infof("response [%s] with cookie[%d]", box.s.GetPeer(), box.s.Cookie)
}

func establishConnection(box *Box) error {
	ipv4, err := box.s.GetPeerIPv4()
	if err != nil {
		return nil
	}
	binary.BigEndian.PutUint32(box.b[8:12], uint32(0))
	binary.BigEndian.PutUint32(box.b[12:16], box.s.ThatSID)
	binary.BigEndian.PutUint16(box.b[22:24], uint16(1))
	binary.BigEndian.PutUint32(box.b[40:44], box.s.ThisSID)
	box.b[48] = ipv4[3]
	box.b[49] = ipv4[2]
	box.b[50] = ipv4[1]
	box.b[51] = ipv4[0]

	binary.BigEndian.PutUint16(box.b[64:66], uint16(srt.HSExtTypeHSRsp))
	binary.BigEndian.PutUint16(box.b[76:78], math.MaxUInt16(box.s.TSBPD.TxDelay, config.GetRx()))
	binary.BigEndian.PutUint16(box.b[78:80], math.MaxUInt16(box.s.TSBPD.RxDelay, config.GetTx()))

	if _, err = box.s.Write(box.b[:80]); err != nil {
		return err
	}
	box.s.Status = session.SConnect
	return nil
}
