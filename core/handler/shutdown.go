package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

type shutdown struct {
	nextHandler srtHandler
}

func NewShutdown() *shutdown {
	h := new(shutdown)
	return h
}

func (h *shutdown) hasNext() bool {
	return h.nextHandler != nil
}

func (h *shutdown) next(next srtHandler) {
	h.nextHandler = next
}

func (h *shutdown) execute(s *session.SRTSession) error {
	if s.CP != nil && s.CP.CType == srt.CTShutdown {
		log.Infof("stream[%s] session shutdown", s.StreamID)
		s.Status = session.SShutdown
		return nil
	} else if h.hasNext() {
		return h.nextHandler.execute(s)
	}
	return errors.New("no handler after shutdown")
}
