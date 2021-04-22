package handler

import (
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

func (d *ackack) execute(box *Box) error {
	if box.s.CP != nil && box.s.CP.CType == srt.CTAckAck && box.s.Status == session.SConnect {
		box.s.ACKNo = box.s.CP.SpecInfo + 1
		rtt := (uint32(time.Now().Unix()) - box.s.ACKTime) << 1
		box.s.RTTDiff = rtt - box.s.RTTTime
		box.s.RTTTime = rtt
		log.Infof("cal new rtt[%d], rttdiff[%d] in ackNo[%d]", box.s.RTTTime, box.s.RTTDiff, box.s.ACKNo-1)
		box.s.CP = nil
		return nil
	} else if d.hasNext() {
		return d.nextHandler.execute(box)
	}
	return errors.New("no handler after ackack")
}
