package srt

import (
	"bytes"
	"encoding/binary"
	"github.com/beleege/gosrt/util/codec"
	"time"
)

// SRT packets are transmitted as UDP payload
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |            SrcPort            |            DstPort            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |              Len              |            ChkSum             |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                          SRT Packet                           +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

// Packet the SRT
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+- SRT Header +-+-+-+-+-+-+-+-+-+-+-+-+-+
// |F|        (Field meaning depends on the packet type)           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |          (Field meaning depends on the packet type)           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           Timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                     Destination Socket ID                     |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                        Packet Contents                        |
// |                  (depends on the packet type)                 +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Packet struct {
	Timestamp uint32 // The timestamp of the packet, in microseconds
	SocketID  uint32 // A fixed-width field providing the SRT socket ID to which a packet should be dispatched
}

// DataPacket structure
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+- SRT Header +-+-+-+-+-+-+-+-+-+-+-+-+-+
// |0|                    Packet Sequence Number                   |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |P P|O|K K|R|                   Message Number                  |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           Timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                     Destination Socket ID                     |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                              Data                             +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type DataPacket struct {
	Packet
	SequenceNum uint32 // The sequential number of the data packet
	PP          uint8  // This field indicates the position of the data packet in the message: (10b) first, (00b) middle, (01b) last, (11b) whole
	O           bool   // Order Flag. Indicates whether the message should be delivered by the receiver in order (1) or not (0)
	KK          uint8  // Key-based Encryption Flag: (00b) not encrypted, (01b) encrypted with even key, (10b) odd key encryption, (11b) for control packet use
	R           bool   // Retransmitted Packet Flag: (0b) first, (1b) retransmitted
	MsgNum      uint32 // The sequential number of consecutive data packets that form a message (see PP field)
	Content     []byte // The payload of the data packet
}

// ControlPacket structure
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+- SRT Header +-+-+-+-+-+-+-+-+-+-+-+-+-+
// |1|         Control Type        |            Subtype            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                   Type-specific Information                   |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           Timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                     Destination Socket ID                     |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+- CIF -+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                   Control Information Field                   +
// |                                                               |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type ControlPacket struct {
	Packet
	CType    uint16 // Control Packet Type
	Subtype  uint16 // This field specifies an additional subtype for specific packets
	SpecInfo uint32 // The use of this field depends on the particular control packet type
	CIF      []byte // The use of this field is defined by the Control Type field of the control packet
}

func (cp *ControlPacket) header(t *time.Time, info uint32, sid uint32) *bytes.Buffer {
	buf := bytes.NewBuffer([]byte{})
	_ = binary.Write(buf, binary.BigEndian, cp.CType|0x8000)
	_ = binary.Write(buf, binary.BigEndian, cp.Subtype)
	_ = binary.Write(buf, binary.BigEndian, info)
	_ = binary.Write(buf, binary.BigEndian, uint32(time.Now().UnixNano()-t.UnixNano()))
	_ = binary.Write(buf, binary.BigEndian, sid)
	return buf
}

func (cp *ControlPacket) Ack(no, sid, seq, rtt, rttDiff, leftMFW, pRate, bandwidth, rRate uint32, t *time.Time) []byte {
	buf := cp.header(t, no, sid)
	_ = binary.Write(buf, binary.BigEndian, seq)
	_ = binary.Write(buf, binary.BigEndian, rtt)
	_ = binary.Write(buf, binary.BigEndian, rttDiff)
	_ = binary.Write(buf, binary.BigEndian, leftMFW)
	_ = binary.Write(buf, binary.BigEndian, pRate)
	_ = binary.Write(buf, binary.BigEndian, bandwidth)
	_ = binary.Write(buf, binary.BigEndian, rRate)
	return buf.Bytes()
}

func (cp *ControlPacket) Shutdown(t *time.Time, sid uint32) []byte {
	buf := cp.header(t, uint32(0), sid)
	_ = binary.Write(buf, binary.BigEndian, uint32(0))
	return buf.Bytes()
}

// HandShakeCIF
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                            Version                            |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |        Encryption Field       |        Extension Field        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                 Initial Packet Sequence Number                |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                 Maximum Transmission Unit Size                |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                    Maximum Flow Window Size                   |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         Handshake Type                        |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                         SRT Socket ID                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           SYN Cookie                          |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                                                               +
// |                                                               |
// +                        Peer IP Address                        +
// |                                                               |
// +                                                               +
// |                                                               |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// |         Extension Type        |        Extension Length       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                                                               |
// +                       Extension Contents                      +
// |                                                               |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
type HandShakeCIF struct {
	Version         uint32 // A base protocol version number.  Currently used values are 4 and 5
	Encryption      uint16 // Block cipher family and key size
	Extension       uint16 // This field is message specific extension related to Handshake Type field
	InitSequenceNum uint32 // The sequence number of the very first data packet to be sent
	MTU             uint32 // This value is typically set      to 1500, which is the default Maximum Transmission Unit (MTU) size  for Ethernet, but can be less
	MFW             uint32 // The value of this field is the maximum number of data packets allowed to be "in flight"
	HType           uint32 // his field indicates the handshake packet type
	SocketID        uint32 // This field holds the ID of the source SRT socket from which a handshake packet is issued
	Cookie          uint32 // Randomized value for processing a handshake
	PeerIP          []byte // IPv4 or IPv6 address of the packet’s sender
	HSExt           []byte // Multi HSExtension
}

type HSExtension struct {
	EType    uint16 // The value of this field is used to process an integrated handshake
	ELength  uint16 // The length of the Extension Contents field in four-byte blocks
	EContent []byte // The payload of the extension, see HSExtTSBPD, HSExtStreamID etc
}

// HSExtTSBPD SRT Extension
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                          SRT Version                          |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           SRT Flags                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |      Receiver TSBPD Delay     |       Sender TSBPD Delay      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type HSExtTSBPD struct {
	Bytes      []byte
	SRTVersion uint32 // SRT library version
	SRTFlags   uint32 // SRT configuration flags
	TxDelay    uint16 // Delay of the Receiver
	RxDelay    uint16 // Delay of the Sender
}

// HSExtStreamID SRT Extension (Stream ID)
type HSExtStreamID struct {
	StreamID string
}

func ParseDPacket(b []byte) *DataPacket {
	p := new(DataPacket)
	b = codec.Decode32u(b, &p.SequenceNum)
	p.PP = (b[0] & 0xC0) >> 6
	p.O = (b[0] & 0x20) > 0
	p.KK = (b[0] & 0x18) >> 3
	p.R = (b[0] & 0x04) > 0
	b = codec.Decode32u(b, &p.MsgNum)
	p.MsgNum &= 0x11FFFFFF
	b = codec.Decode32u(b, &p.Timestamp)
	b = codec.Decode32u(b, &p.SocketID)
	p.Content = b[:]
	return p
}

func ParseCPacket(b []byte) *ControlPacket {
	p := new(ControlPacket)
	b = codec.Decode16u(b, &p.CType)
	p.CType &= 0x7FFF
	b = codec.Decode16u(b, &p.Subtype)
	b = codec.Decode32u(b, &p.SpecInfo)
	b = codec.Decode32u(b, &p.Timestamp)
	b = codec.Decode32u(b, &p.SocketID)
	p.CIF = b[:]
	return p
}

func ParseHCIF(b []byte) *HandShakeCIF {
	h := new(HandShakeCIF)
	b = codec.Decode32u(b, &h.Version)
	b = codec.Decode16u(b, &h.Encryption)
	b = codec.Decode16u(b, &h.Extension)
	b = codec.Decode32u(b, &h.InitSequenceNum)
	b = codec.Decode32u(b, &h.MTU)
	b = codec.Decode32u(b, &h.MFW)
	b = codec.Decode32u(b, &h.HType)
	b = codec.Decode32u(b, &h.SocketID)
	b = codec.Decode32u(b, &h.Cookie)
	h.PeerIP = b[:16]
	h.HSExt = b[16:]
	return h
}

func ParseHExtension(b []byte) *HSExtTSBPD {
	h := new(HSExtTSBPD)
	h.Bytes = b
	b = codec.Decode32u(b, &h.SRTVersion)
	b = codec.Decode32u(b, &h.SRTFlags)
	b = codec.Decode16u(b, &h.TxDelay)
	b = codec.Decode16u(b, &h.RxDelay)
	return h
}
