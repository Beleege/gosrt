package pool

import (
	"sync"

	"github.com/beleege/gosrt/util/log"
	"github.com/pkg/errors"
)

type Task func(args ...interface{}) error

type Job interface {
	GetID() string
	GetTask() Task
}

type worker struct {
	id    string
	queue chan Job
}

func (w *worker) submit(j Job) {
	if j == nil {
		return
	}

	w.queue <- j
}

func (w *worker) start() error {
	for j := range w.queue {
		if err := j.GetTask()(); err != nil {
			log.Errorf("job[%s] execute fail: %s", j.GetID(), err.Error())
			return err
		}
	}
	return nil
}

type Pool struct {
	size int
	used int
	m    *sync.Map
}

func NewFixedSizePool(s int) *Pool {
	p := new(Pool)
	p.size = s
	p.m = &sync.Map{}
	return p
}

func (p *Pool) Execute(j Job) error {
	key := j.GetID()
	if len(key) == 0 {
		return errors.New("key is empty")
	}

	var w *worker
	v, ok := p.m.Load(key)
	if !ok {
		if p.used+1 > p.size {
			return errors.Errorf("overcome capacity[%d]", p.size)
		}
		p.used++

		w = new(worker)
		w.id = j.GetID()
		w.queue = make(chan Job)
		p.m.Store(key, w)
		go w.start()
	} else {
		w = v.(*worker)
	}
	w.submit(j)

	return nil
}

func (p *Pool) Remove(key string) {
	v, ok := p.m.LoadAndDelete(key)
	if ok {
		p.used--
		close(v.(*worker).queue)
	}
}

func (p *Pool) Clear() {
	p.m.Range(func(k, v interface{}) bool {
		close(v.(*worker).queue)
		return true
	})
	p.m = nil
}
