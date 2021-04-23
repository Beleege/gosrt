package window

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beleege/gosrt/protocol/srt"
)

type NoLossAction func(seq uint32)

type underlay struct {
	// update lock
	mu sync.Mutex
	// drop task condition
	cond *sync.Cond
	// window size
	size int
	// used window size
	used int
	// maintain list
	list []node
	// dirty list
	dirty []node
	// map to list
	dict map[int]*node
	// initial seq no
	first uint32
	// last seq no
	last uint32
	// initial timestamp
	ts int64
	// pkg add counter
	counter uint64
	// pkg add event
	eventChan chan struct{}
	// loss channel
	lossChan chan []LossRange
	// batch channel
	batchChan chan []*srt.DataPacket
	// loss monitor start
	lossMon int32
	// no loss pkg in window action, normally for ack
	act NoLossAction
}

type node struct {
	pkg *srt.DataPacket
	// create time
	t int64
	// is pkg loss
	loss bool
}

type LossRange struct {
	Start uint32
	End   uint32
}

func NewWindow(size int, act NoLossAction) *underlay {
	p := new(underlay)
	p.size = size
	p.list = make([]node, size*2)
	p.dirty = p.list[:]
	p.dict = make(map[int]*node)
	p.eventChan = make(chan struct{})
	p.lossChan = make(chan []LossRange, 2048)
	p.batchChan = make(chan []*srt.DataPacket, 2048)

	p.mu = sync.Mutex{}
	p.cond = sync.NewCond(&p.mu)
	p.act = act

	go p.onPkgAdd()

	return p
}

func (u *underlay) TS() int64 {
	return u.ts
}

func (u *underlay) ListenLoss() chan []LossRange {
	return u.lossChan
}

func (u *underlay) ListenBatch() chan []*srt.DataPacket {
	return u.batchChan
}

func (u *underlay) Append(p *srt.DataPacket) bool {
	now := time.Now().Unix()
	u.cond.L.Lock()
	defer u.cond.L.Unlock()

	if u.used >= u.size {
		return false
	}
	u.counter++

	// pkg in event
	u.eventChan <- struct{}{}

	if u.used == 0 {
		u.ts = now
		u.first = p.SequenceNum
		u.last = p.SequenceNum
		u.dirty[0].t = u.ts
		u.dirty[0].pkg = p
		u.used++
	} else {
		pos := p.SequenceNum - u.first
		if pos < 0 {
			return true
		} else if int(pos) < u.used-1 {
			u.dirty[pos].pkg = p
			u.dirty[pos].loss = false

			delete(u.dict, int(pos))
		} else if int(pos) == u.used-1 {
			u.dirty[u.used-1].loss = false
			u.dirty[u.used-1].t = now
			u.dirty[u.used-1].pkg = p
			u.last = p.SequenceNum
			u.used++
		} else {
			for i, n := u.used, int(pos); i < n; i++ {
				u.dirty[i].loss = true
				u.dirty[i].pkg = nil
				u.dirty[i].t = now
				u.dict[i] = &u.dirty[i]
				u.used++
			}
			u.dirty[int(pos)].loss = false
			u.dirty[int(pos)].t = now
			u.dirty[int(pos)].pkg = p
			u.last = p.SequenceNum
			u.used++

			// start loss monitor
			go u.lossMonitor()
		}
	}
	return true
}

func (u *underlay) IsFull() bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.used == u.size && len(u.dict) == 0
}

func (u *underlay) Loss() []LossRange {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.dict) > 0 {
		nos := make([]int, 0, len(u.dict))
		for i := range u.dict {
			nos = append(nos, i)
		}
		if len(nos) == 1 {
			return []LossRange{{Start: uint32(nos[0]) + u.first, End: uint32(nos[0]) + u.first}}
		}

		sort.Ints(nos)
		ranges := make([]LossRange, 0, 4)
		no := nos[0]
		for idx := 1; idx <= len(nos); idx++ {
			if idx < len(nos) && nos[idx]-1 == nos[idx-1] {
				continue
			}
			ranges = append(ranges, LossRange{Start: uint32(no) + u.first, End: uint32(nos[idx-1]) + u.first})
			if idx < len(nos) {
				no = nos[idx]
			}
		}
		return ranges
	}
	return make([]LossRange, 0)
}

// warn: need lock protect
func (u *underlay) reset() {
	u.ts = 0
	u.first = 0
	u.dirty = u.list[:]
	u.used = 0
	for i := range u.dirty {
		u.dirty[i].pkg = nil
		u.dirty[i].t = 0
		u.dirty[i].loss = false
	}
	if len(u.dict) > 0 {
		u.dict = make(map[int]*node)
	}
}

func (u *underlay) lossMonitor() {
	if !atomic.CompareAndSwapInt32(&u.lossMon, 0, 1) {
		return
	}
	defer func() {
		u.lossMon = 0
	}()

	timer := time.NewTimer(120 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			// check loss
			loss := u.Loss()
			if len(loss) > 0 {
				u.lossChan <- loss
				timer.Reset(120 * time.Millisecond)
			} else {
				return
			}
		}
	}
}

func (u *underlay) onPkgAdd() {
	timer := time.NewTimer(10 * time.Millisecond)

	for {
	LOOP:
		select {
		case <-timer.C:
			timer.Reset(10 * time.Millisecond)
		case <-u.eventChan:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(10 * time.Millisecond)
			goto LOOP
		}

		u.cond.L.Lock()
		// check drop
		now := time.Now().Unix()
		offset := 0
		for i := range u.dirty {
			// a packet timestamp is older than 125% of the SRT latency
			if u.dirty[i].loss && now-u.dirty[i].t >= 150 {
				offset = int(math.Max(float64(offset), float64(i)))
			}
		}
		if offset > 0 {
			u.dirty = u.dirty[offset+1:]
			u.dict = make(map[int]*node)
			for i := range u.dirty {
				if u.dirty[i].loss {
					u.dict[i] = &u.dirty[i]
				}
			}
		}
		// no loss and have some pkgs，delivery them
		if len(u.dict) == 0 {
			// do no loss action
			go u.act(u.last)

			arr := make([]*srt.DataPacket, 0, len(u.dirty))
			for i := range u.dirty {
				arr = append(arr, u.dirty[i].pkg)
			}
			if len(arr) > 0 {
				u.batchChan <- arr
				u.reset()
			}
		}
		u.cond.L.Unlock()
	}
}
