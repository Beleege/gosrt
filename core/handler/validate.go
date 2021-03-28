package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/pkg/errors"
)

const (
	_maxPacketBytes = 1500
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
	if size := len(s.Data); size > _maxPacketBytes {
		return errors.Errorf("package size[%d] is illegal", len(s.Data))
	} else if v.hasNext() {
		return v.nextHandler.execute(s)
	}
	return errors.New("no handler after validator")
}
