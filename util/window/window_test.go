package window

import (
	"github.com/beleege/gosrt/protocol/srt"
	"testing"
)

func TestWindow(t *testing.T) {
	win := NewWindow(10)
	win.Append(&srt.DataPacket{SequenceNum: 100})
	win.Append(&srt.DataPacket{SequenceNum: 105})
	win.Append(&srt.DataPacket{SequenceNum: 106})
	win.Append(&srt.DataPacket{SequenceNum: 108})
	t.Logf("lost seqs: %+v", win.Miss())
	win.Append(&srt.DataPacket{SequenceNum: 104})
	win.Append(&srt.DataPacket{SequenceNum: 107})
	t.Logf("lost seqs: %+v", win.Miss())
	win.Reset()
	t.Logf("win is %+v", win)
}

func TestFull(t *testing.T) {
	win := NewWindow(3)
	win.Append(&srt.DataPacket{SequenceNum: 100})
	win.Append(&srt.DataPacket{SequenceNum: 102})
	t.Logf("win is full: %v", win.IsFull())
	t.Logf("lost seqs: %+v", win.Miss())
}
