package handler

import (
	"encoding/binary"
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

func (d *dataStream) execute(box *Box) error {
	if box.s.DP != nil {
		var q []*srt.DataPacket
		v, ok := _queueMap.Load(box.s.GetPeer())
		if !ok {
			q = make([]*srt.DataPacket, 0, box.s.MFW)
		} else {
			q = v.([]*srt.DataPacket)
		}

		q = append(q, box.s.DP)
		_queueMap.Store(box.s.GetPeer(), q)
		log.Infof("received data in queue[%d] with seqNo: %d", len(q), binary.BigEndian.Uint32(box.b[:4]))
		//if len(q)%24 == 0 {
		//	return ack(s, q)
		//} else if len(q) >= int(s.MFW) {
		//	q = q[0:0]
		//}
		return nil
	} else if d.hasNext() {
		return d.nextHandler.execute(box)
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
