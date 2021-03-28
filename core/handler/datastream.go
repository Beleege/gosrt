package handler

import (
	"sync"

	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

var _queueMap = sync.Map{}

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

func (d *dataStream) execute(s *session.SRTSession) error {
	if s.DP != nil {
		var q []*srt.DataPacket
		v, ok := _queueMap.Load(s.GetPeer())
		if !ok {
			q = make([]*srt.DataPacket, 0, s.MFW)
		} else {
			q = v.([]*srt.DataPacket)
		}

		q = append(q, s.DP)
		_queueMap.Store(s.GetPeer(), q)
		log.Infof("received data in queue[%d] addrt: %p", len(q), &q)
		if len(q)%50 == 0 {
			return ack(s, q)
		} else if len(q) >= int(s.MFW) {
			q = q[0:0]
		}
		return nil
	} else if d.hasNext() {
		return d.nextHandler.execute(s)
	}

	return errors.New("no handler after dataStream")
}

func ack(s *session.SRTSession, q []*srt.DataPacket) error {
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
	leftMFW := s.MFW - uint32(len(q))
	pRate := 1000
	bandwidth := 1000
	rRate := 1000
	_, _ = s.Write(cp.Ack(ackNo, s.ThatSID, s.SendNo+1, rtt, rttDiff, leftMFW, uint32(pRate), uint32(bandwidth), uint32(rRate), &s.OpenTime))
	return nil
}
