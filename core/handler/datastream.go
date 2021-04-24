package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

type dataStream struct {
	nextHandler srtHandler
}

func NewDataStream() *dataStream {
	d := new(dataStream)
	return d
}

func (d *dataStream) hasNext() bool {
	return d.nextHandler != nil
}

func (d *dataStream) next(next srtHandler) {
	d.nextHandler = next
}

func (d *dataStream) execute(box *Box) error {
	if box.s.DP != nil {
		box.s.RecWin.Append(box.s.DP)
		box.s.AddACKAction(ack)
		return nil
	} else if d.hasNext() {
		return d.nextHandler.execute(box)
	}

	return errors.New("no handler after dataStream")
}

func ack(s *session.SRTSession, seq uint32) {
	log.Infof("fire ack to %s", s.StreamID)
	cp := new(srt.ControlPacket)
	cp.CType = srt.CTAck

	ackNo := s.ACKNo
	if ackNo <= 0 {
		ackNo = 1
	}
	rtt := s.RTTTime
	if rtt <= 0 {
		rtt = 100000
	}
	rttDiff := s.RTTDiff
	if rttDiff <= 0 {
		rttDiff = 50000
	}
	leftMFW := s.MFW - s.RecWin.Len()
	pRate := 1000
	bandwidth := 1000
	rRate := 1000
	_, _ = s.Write(cp.Ack(ackNo, s.ThatSID, seq+1, rtt, rttDiff, leftMFW, uint32(pRate), uint32(bandwidth), uint32(rRate), &s.OpenTime))
}
