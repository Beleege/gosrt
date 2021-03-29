package handler

import (
	"encoding/hex"

	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
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
		log.Debugf("-----------------------------------------------")
		log.Debugf("binary data:\n%s", hex.Dump(s.Data))
		pkg := srt.ParseCPacket(s.Data)
		log.Debugf("control pkg type is %d", pkg.CType)
		if pkg.CType == srt.CTHandShake {
			cif := srt.ParseHCIF(pkg.CIF)
			s.SetCP(pkg, cif)
		} else {
			s.CP = pkg
		}
	} else if t == srt.PTypeData {
		if s.Status != session.SConnect {
			// TODO clear session
			return errors.Errorf("session is not connected")
		}
		pkg := srt.ParseDPacket(s.Data)
		s.SetDP(pkg)
	}
	if d.hasNext() {
		return d.nextHandler.execute(s)
	}
	return errors.New("no handler after decoder")
}
