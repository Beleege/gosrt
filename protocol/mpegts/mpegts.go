package mpegts

import (
	"github.com/beleege/gosrt/util/codec"
)

const (
	_syncCode     = 0x47
	TSPackageSize = 188
)

type Header struct {
	Sync             uint8
	Err              uint8
	PayloadUnitStart uint8
	Prio             uint8
	PID              uint16
	Scra             uint8
	Adaptation       uint8
	Counter          uint8
}

func ParseHeader(b []byte) *Header {
	h := new(Header)
	b = codec.Decode8u(b, &h.Sync)
	h.Err = (b[0] & 0x80) >> 7
	h.PayloadUnitStart = (b[0] & 0x40) >> 6
	h.Prio = (b[0] & 0x20) >> 5
	b = codec.Decode16u(b, &h.PID)
	h.PID &= 0x1F
	h.Scra = (b[0] & 0xC0) >> 6
	h.Adaptation = (b[0] & 0x30) >> 4
	h.Counter = b[0] & 0x0F
	return h
}

func ExtractPCR(b []byte) (float64, bool) {
	if b[0] != _syncCode {
		return 0, false
	}
	pcrAdaptationFlag := (b[3] & 0x30) >> 4
	if pcrAdaptationFlag != 2 && pcrAdaptationFlag != 3 {
		return 0, false
	}
	if b[4] == 0 {
		// adaptation_field_length is zero
		return 0, false
	}
	pcrFlag := b[5] & 0x10
	if pcrFlag == 0 {
		return 0, false
	}
	// there's a PCR
	pcrBaseHigh := int64(b[6])<<24 | int64(b[7])<<16 | int64(b[8])<<8 | int64(b[9])
	clock := float64(pcrBaseHigh / 45000.0)
	if (b[10] & 0x80) != 0 {
		clock += 1.0 / 90000
	}
	pcrExt := int64(b[10]&0x01)<<8 | int64(b[11])
	clock += float64(pcrExt / 27000000.0)
	return clock, true
}
