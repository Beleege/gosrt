package srt

import (
	"encoding/binary"
)

const ()

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

// The structure of the SRT packet
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

// Data packet structure
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
	O           uint8  // Order Flag. Indicates whether the message should be delivered by the receiver in order (1) or not (0)
	KK          uint8  // Key-based Encryption Flag: (00b) not encrypted, (01b) encrypted with even key, (10b) odd key encryption, (11b) for control packet use
	R           uint8  // Retransmitted Packet Flag: (0b) first, (1b) retransmitted
	MsgNum      uint8  // The sequential number of consecutive data packets that form a message (see PP field)
	data        []byte // The payload of the data packet
}

// Control packet structure
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
	SpecInfo []byte // The use of this field depends on the particular control packet type
	CIF      []byte // The use of this field is defined by the Control Type field of the control packet
}

// CIF of HandShake
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

// SRT Extension (SRT_CMD_HSREQ & SRT_CMD_HSRSP)
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
	SRTVersion uint32 // SRT library version
	SRTFlags   uint32 // SRT configuration flags
	ThisDelay  uint16 // Delay of the sender
	ThatDelay  uint16 // Delay of the peer
}

// SRT Extension (Stream ID)
type HSExtStreamID struct {
	StreamID string
}

func ParseCPacket(b []byte) *ControlPacket {
	b[0] = b[0] & 0x7F
	p := new(ControlPacket)
	p.CType = binary.BigEndian.Uint16(b[:2])
	p.Subtype = binary.BigEndian.Uint16(b[2:4])
	p.SpecInfo = b[4:8]
	p.Timestamp = binary.BigEndian.Uint32(b[8:12])
	p.SocketID = binary.BigEndian.Uint32(b[12:16])
	p.CIF = b[16:]
	return p
}

func ParseHCIF(b []byte) *HandShakeCIF {
	h := new(HandShakeCIF)
	h.Version = binary.BigEndian.Uint32(b[:4])
	h.Encryption = binary.BigEndian.Uint16(b[4:6])
	h.Extension = binary.BigEndian.Uint16(b[6:8])
	h.InitSequenceNum = binary.BigEndian.Uint32(b[8:12])
	h.MTU = binary.BigEndian.Uint32(b[12:16])
	h.MFW = binary.BigEndian.Uint32(b[16:20])
	h.HType = binary.BigEndian.Uint32(b[20:24])
	h.SocketID = binary.BigEndian.Uint32(b[24:28])
	h.Cookie = binary.BigEndian.Uint32(b[28:32])
	h.PeerIP = b[32:48]
	if len(b) > 48 {
		h.HSExt = b[48:]
	}
	return h
}

func ParseHExtension(b []byte) *HSExtTSBPD {
	h := new(HSExtTSBPD)
	h.SRTVersion = binary.BigEndian.Uint32(b[:4])
	h.SRTFlags = binary.BigEndian.Uint32(b[4:8])
	h.ThatDelay = binary.BigEndian.Uint16(b[8:10])
	h.ThisDelay = binary.BigEndian.Uint16(b[10:12])
	return h
}
