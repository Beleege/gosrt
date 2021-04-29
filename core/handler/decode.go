package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
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

func (d *decoder) execute(box *Box) error {
	t := box.b[:1][0] >> 7
	if t == srt.PTypeControl {
		//log.Debugf("-----------------------------------------------")
		//log.Debugf("binary data:\n%s", hex.Dump(s.Data))
		pkg := srt.ParseCPacket(box.b)
		//log.Debugf("control pkg type is %d", pkg.CType)
		if pkg.CType == srt.CTHandShake {
			cif := srt.ParseHCIF(pkg.CIF)
			box.s.SetCP(pkg, cif)
		} else {
			box.s.CP = pkg
		}
	} else if t == srt.PTypeData {
		if box.s.Status.Load().(int) != session.SConnect {
			// TODO clear session
			return errors.Errorf("session is not connected")
		}
		pkg := srt.ParseDPacket(box.b)
		box.s.SetDP(pkg)
	}
	if d.hasNext() {
		return d.nextHandler.execute(box)
	}
	return errors.New("no handler after decoder")
}
