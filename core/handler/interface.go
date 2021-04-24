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

type Box struct {
	s *session.SRTSession
	b []byte
}

func NewBox(s *session.SRTSession, b []byte) *Box {
	box := new(Box)
	box.s = s
	box.b = b
	return box
}

var Queue = make(chan *Box, 1024)

var taskPool *pool.Pool

type srtHandler interface {
	hasNext() bool
	next(h srtHandler)
	execute(s *Box) error
}

type srtContext struct {
	handler srtHandler
	box     *Box
}

func (c *srtContext) GetID() string {
	return strconv.Itoa(int(c.box.s.ThatSID))
}

func (c *srtContext) GetTask() pool.Task {
	return func(args ...interface{}) error {
		if err := c.handler.execute(c.box); err != nil {
			log.Errorf("handle session fail: %s", err.Error())
			closeConnect(c.box)
		}
		return nil
	}
}

func Task() {
	taskPool = pool.NewFixedSizePool(config.GetPoolSize())
	defer taskPool.Clear()

	chain := wrap()
	for box := range Queue {
		ctx := &srtContext{handler: chain, box: box}
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

func closeConnect(box *Box) {
	s := box.s
	if s.Status.Load().(int) == session.SShutdown {
		return
	}
	p := new(srt.ControlPacket)
	p.CType = srt.CTShutdown
	p.Subtype = uint16(0)
	p.SpecInfo = 0
	p.Timestamp = uint32(time.Now().UnixNano() - s.OpenTime.UnixNano())
	p.SocketID = s.ThatSID

	_, _ = s.Write(p.Shutdown(&s.OpenTime, s.ThatSID))
}
