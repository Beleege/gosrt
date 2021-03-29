package handler

import (
	"encoding/binary"
	"time"

	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

type ackack struct {
	nextHandler srtHandler
}

func NewAckAck() *ackack {
	d := new(ackack)
	return d
}

func (d *ackack) hasNext() bool {
	return d.nextHandler != nil
}

func (d *ackack) next(next srtHandler) {
	d.nextHandler = next
}

func (d *ackack) execute(s *session.SRTSession) error {
	if s.CP != nil && s.CP.CType == srt.CTAckAck && s.Status == session.SConnect {
		s.ACKNo = binary.BigEndian.Uint32(s.CP.SpecInfo) + 1
		rtt := (uint32(time.Now().Unix()) - s.ACKTime) << 1
		s.RTTDiff = rtt - s.RTTTime
		s.RTTTime = rtt
		log.Infof("cal new rtt[%d], rttdiff[%d] in ackNo[%d]", s.RTTTime, s.RTTDiff, s.ACKNo-1)
		s.CP = nil
		return nil
	} else if d.hasNext() {
		return d.nextHandler.execute(s)
	}
	return errors.New("no handler after ackack")
}
