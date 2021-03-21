package srt

import (
	"testing"
)

func TestParseCPacket(t *testing.T) {
	b := []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x8d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02, 0x4a, 0x5d, 0x18, 0xe4, 0x00, 0x00, 0x05, 0xdc, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x01, 0x06, 0x5f, 0x4e, 0x8f, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p := ParseCPacket(b)
	t.Logf("%+v", *p)
}

func TestParseHCIF(t *testing.T) {
	b := []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x8d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02, 0x4a, 0x5d, 0x18, 0xe4, 0x00, 0x00, 0x05, 0xdc, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x01, 0x06, 0x5f, 0x4e, 0x8f, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p := ParseCPacket(b)
	h := ParseHCIF(p.CIF)
	t.Logf("%+v", h)
}
