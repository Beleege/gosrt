package handler

import (
	"github.com/pkg/errors"
)

const (
	_maxPacketBytes = 1400
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

func (v *validator) execute(box *Box) error {
	if size := len(box.b); size > _maxPacketBytes {
		return errors.Errorf("package size[%d] is illegal", len(box.b))
	} else if v.hasNext() {
		return v.nextHandler.execute(box)
	}
	return errors.New("no handler after validator")
}
