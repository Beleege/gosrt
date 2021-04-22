package codec

import "encoding/binary"

func Encode8u(p []byte, c byte) []byte {
	p[0] = c
	return p[1:]
}

func Decode8u(p []byte, c *byte) []byte {
	*c = p[0]
	return p[1:]
}

func Encode16u(p []byte, w uint16) []byte {
	binary.BigEndian.PutUint16(p, w)
	return p[2:]
}

func Decode16u(p []byte, w *uint16) []byte {
	*w = binary.BigEndian.Uint16(p)
	return p[2:]
}

func Encode32u(p []byte, l uint32) []byte {
	binary.BigEndian.PutUint32(p, l)
	return p[4:]
}

func Decode32u(p []byte, l *uint32) []byte {
	*l = binary.BigEndian.Uint32(p)
	return p[4:]
}
