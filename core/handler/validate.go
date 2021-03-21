package handler

import (
	"github.com/beleege/gosrt/core/session"
)

const (
	_minPacketBytes = 96
)

type validator struct {
	nextHandler srtHandler
}

func NewValidator() *validator {
	v := new(validator)
	return v
}

func (v *validator) hasNext() bool {
	return v.nextHandler != nil
}

func (v *validator) next(next srtHandler) {
	v.nextHandler = next
}

func (v *validator) execute(s *session.SRTSession) error {
	if len(s.Data) < _minPacketBytes {
		if _, err := s.Write(s.Data); err != nil {
			return err
		}
	} else if v.hasNext() {
		return v.nextHandler.execute(s)
	}
	return nil
}
