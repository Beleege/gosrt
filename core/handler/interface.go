package handler

import (
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/util/log"
)

var Queue = make(chan *session.SRTSession)

type srtHandler interface {
	hasNext() bool
	next(h srtHandler)
	execute(s *session.SRTSession) error
}

func Task() {
	defer close(Queue)
	chain := wrap()

	for s := range Queue {
		go doWork(chain, s)
	}
}

func doWork(h srtHandler, s *session.SRTSession) {
	if err := h.execute(s); err != nil {
		log.Error("handle session fail: %s", err.Error())
	}
}

func wrap() srtHandler {
	handlers := selectHandlers()
	for i := len(handlers) - 2; i >= 0; i-- {
		handlers[i].next(handlers[i+1])
	}
	return handlers[0]
}

func selectHandlers() []srtHandler {
	list := make([]srtHandler, 0, 3)
	list = append(list, NewValidator())
	list = append(list, NewDecoder())
	list = append(list, NewGreeter())
	return list
}
