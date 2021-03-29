package session

import (
	"encoding/binary"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/beleege/gosrt/protocol/srt"
	"github.com/beleege/gosrt/util/log"
)

const (
	SNew       = 0
	SOpen      = 1
	SSetCookie = 2
	SRepeat    = 3
	SConnect   = 4
	SIllegal   = 0xFFFFFFFE
	SShutdown  = 0xFFFFFFFF
)

type SRTSession struct {
	conn     net.PacketConn
	peer     net.Addr
	OpenTime time.Time
	ACKNo    uint32
	ACKTime  uint32
	RTTTime  uint32
	RTTDiff  uint32

	CP   *srt.ControlPacket
	DP   *srt.DataPacket
	Data []byte

	SendNo   uint32
	MTU      uint32
	MFW      uint32
	ThisSID  uint32
	ThatSID  uint32
	Cookie   uint32
	StreamID string
	TSBPD    *srt.HSExtTSBPD
	// TODO condition compete
	Status uint32
}

func (s *SRTSession) Write(b []byte) (n int, err error) {
	return s.conn.WriteTo(b, s.peer)
}

func NewSRTSession(c net.PacketConn, a net.Addr, b []byte) *SRTSession {
	s := new(SRTSession)
	s.conn = c
	s.peer = a
	s.Data = b
	s.OpenTime = time.Now()
	s.ThisSID = rand.New(rand.NewSource(s.OpenTime.UnixNano())).Uint32()
	s.Status = SNew
	return s
}

func (s *SRTSession) SetDP(pkg *srt.DataPacket) {
	s.DP = pkg
	s.SendNo = pkg.SequenceNum
}

func (s *SRTSession) SetCP(pkg *srt.ControlPacket, cif *srt.HandShakeCIF) {
	s.CP = pkg
	s.parseHSExtension(cif.HSExt)
	if cif.Version == srt.HSv4 {
		s.SendNo = cif.InitSequenceNum
		s.MTU = cif.MTU
		s.MFW = cif.MFW
		s.ThatSID = cif.SocketID
		if cif.HType == srt.HSTypeInduction {
			s.Status = SOpen
		}
	} else if cif.Version == srt.HSv5 {
		if cif.Cookie != s.Cookie {
			log.Errorf("cookie[%d] is not match", cif.Cookie)
			s.Status = SIllegal
			return
		}
		if cif.HType == srt.HSTypeConclusion {
			s.Status = SRepeat
			s.ACKTime = uint32(time.Now().Unix())
		}
	}
}

func (s *SRTSession) parseHSExtension(b []byte) {
	if len(b) == 0 {
		return
	}

	exts := parseMultiExt(b)
	if len(exts) == 0 {
		return
	}

	for _, ext := range exts {
		switch ext.EType {
		case 0:
			return
		case srt.HSExtTypeHSReq:
			fallthrough
		case srt.HSExtTypeHSRsp:
			s.TSBPD = srt.ParseHExtension(ext.EContent[4:])
		case srt.HSExtTypeSID:
			s.StreamID = string(ext.EContent[4:])
		}
	}
}

func parseMultiExt(b []byte) []*srt.HSExtension {
	exts := make([]*srt.HSExtension, 0, 2)
	idx := 0
	size := len(b)
	for idx < size {
		hse := new(srt.HSExtension)

		hse.EType = binary.BigEndian.Uint16(b[idx : idx+2])
		idx = idx + 2

		hse.ELength = binary.BigEndian.Uint16(b[idx : idx+2])
		idx = idx + 2
		end := idx + int(hse.ELength)*4

		hse.EContent = b[idx:end]
		exts = append(exts, hse)

		idx = end
	}
	return exts
}

func (s *SRTSession) GetPeerIPv4() (*[4]byte, error) {
	ip := strings.Split(s.peer.String(), ":")[0]
	bytes := new([4]byte)
	for i, a := range strings.Split(ip, ".") {
		n, err := strconv.ParseUint(a, 10, 8)
		if err != nil {
			return bytes, err
		}
		bytes[i] = byte(n)
	}
	return bytes, nil
}

func (s *SRTSession) GetPeer() string {
	return s.peer.String()
}
