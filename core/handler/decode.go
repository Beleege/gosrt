package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
)

type decoder struct {
	nextHandler srtHandler
}

func NewDecoder() *decoder {
	d := new(decoder)
	return d
}

func (d *decoder) hasNext() bool {
	return d.nextHandler != nil
}

func (d *decoder) next(next srtHandler) {
	d.nextHandler = next
}

func (d *decoder) execute(s *session.SRTSession) error {
	t := s.Data[:1][0] >> 7
	if t == srt.PTypeControl {
		p := srt.ParseCPacket(s.Data)
		if p.CType == srt.CTHandShake {
			cif := srt.ParseHCIF(s.CP.CIF)
			s.SetCP(p, cif)
		}
	} else if t == srt.PTypeData {

	}
	return nil
}
