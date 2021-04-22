package window

import (
	"sort"
	"time"

	"github.com/beleege/gosrt/protocol/srt"
)

type underlay struct {
	// linked list
	list []node
	// map to list
	dict map[uint32]*node
	// initial seq no
	first uint32
	// initial timestamp
	ts int64
}

type node struct {
	pkg *srt.DataPacket
	// create time
	t uint32
	// is pkg miss
	miss bool
}

type LossRange struct {
	Start uint32
	End   uint32
}

func NewWindow(size int) *underlay {
	p := new(underlay)
	p.list = make([]node, 0, size)
	p.dict = make(map[uint32]*node)
	return p
}

func (u *underlay) TS() int64 {
	return u.ts
}

func (u *underlay) Append(p *srt.DataPacket) bool {
	size := uint32(len(u.list))
	if size == uint32(cap(u.list)) {
		return false
	}
	if size == 0 {
		u.ts = time.Now().Unix()
		u.first = p.SequenceNum
		u.list = append(u.list, node{pkg: p})
	} else {
		pos := p.SequenceNum - u.first
		if pos < 0 {
			return true
		} else if pos < size {
			u.list[pos].pkg = p
			delete(u.dict, pos)
		} else if pos >= size && pos < uint32(cap(u.list)) {
			left := pos - size
			for i := uint32(0); i < left; i++ {
				dummy := node{}
				u.list = append(u.list, dummy)
				u.dict[size+i] = &dummy
			}
			u.list = append(u.list, node{pkg: p})
		} else {
			return false
		}
	}
	return true
}

func (u *underlay) IsFull() bool {
	return len(u.list) == cap(u.list) && len(u.dict) == 0
}

func (u *underlay) Miss() []LossRange {
	if len(u.dict) > 0 {
		nos := make([]int, 0, len(u.dict))
		for i := range u.dict {
			nos = append(nos, int(i))
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

func (u *underlay) Reset() {
	u.ts = time.Now().Unix()
	u.first = 0
	u.list = u.list[0:0]
	if len(u.dict) > 0 {
		for k := range u.dict {
			delete(u.dict, k)
		}
	}
}
