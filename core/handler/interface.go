package handler

import (
	"strconv"
	"time"

	"github.com/beleege/gosrt/config"
	"github.com/beleege/gosrt/core/session"
	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
	"github.com/beleege/gosrt/util/pool"
)

var Queue = make(chan *session.SRTSession)

var taskPool *pool.Pool

type srtHandler interface {
	hasNext() bool
	next(h srtHandler)
	execute(s *session.SRTSession) error
}

type srtContext struct {
	h srtHandler
	s *session.SRTSession
}

func (c *srtContext) GetID() string {
	return strconv.Itoa(int(c.s.ThatSID))
}

func (c *srtContext) GetTask() pool.Task {
	return func(args ...interface{}) error {
		if err := c.h.execute(c.s); err != nil {
			log.Errorf("handle session fail: %s", err.Error())
			closeConnect(c.s)
		}
		return nil
	}
}

func Task() {
	taskPool = pool.NewFixedSizePool(config.GetPoolSize())
	defer taskPool.Clear()

	chain := wrap()
	for s := range Queue {
		ctx := &srtContext{h: chain, s: s}
		if err := taskPool.Execute(ctx); err != nil {
			log.Errorf("task execute fail: %s", err.Error())
			taskPool.Remove(ctx.GetID())
		}
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
	list := make([]srtHandler, 0, 6)
	list = append(list, NewValidator())
	list = append(list, NewDecoder())
	list = append(list, NewAckAck())
	list = append(list, NewShutdown())
	list = append(list, NewHandshake())
	list = append(list, NewDataStream())
	return list
}

func closeConnect(s *session.SRTSession) {
	if s.Status == session.SShutdown {
		return
	}
	p := new(srt.ControlPacket)
	p.CType = srt.CTShutdown
	p.Subtype = uint16(0)
	p.SpecInfo = []byte{0x00, 0x00, 0x00, 0x00}
	p.Timestamp = uint32(time.Now().UnixNano() - s.OpenTime.UnixNano())
	p.SocketID = s.ThatSID

	_, _ = s.Write(p.Shutdown(&s.OpenTime, s.ThatSID))
}
