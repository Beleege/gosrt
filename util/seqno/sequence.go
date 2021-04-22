package seqno

import (
	"math"
	"math/rand"
)

const (
	_maxOffset     = 0x3FFFFFFF
	_maxSequenceNo = 0x7FFFFFFF
)

func Compare(n1, n2 uint32) uint32 {
	if math.Abs(float64(n1-n2)) < _maxOffset {
		return n1 - n2
	}
	return n2 - n1
}

func Length(n1, n2 uint32) uint32 {
	if n1 < n2 {
		return n2 - n1 + 1
	}
	return n2 - n1 + _maxSequenceNo + 2
}

func SeqOffset(n1, n2 uint32) uint32 {
	if math.Abs(float64(n1-n2)) < _maxOffset {
		return n2 - n1
	}
	if n1 < n2 {
		return n2 - n1 - _maxSequenceNo - 1
	}
	return n2 - n1 + _maxSequenceNo + 1
}

func Increment(n uint32) uint32 {
	if n == _maxSequenceNo {
		return 0
	}
	return n + 1
}

func Decrement(n uint32) uint32 {
	if n == 0 {
		return _maxSequenceNo
	}
	return n - 1
}

func Random() uint32 {
	return (rand.Uint32() % _maxOffset) + 1
}
