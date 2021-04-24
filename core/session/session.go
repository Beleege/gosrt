package session

import (
	"container/list"
	"github.com/beleege/gosrt/util/codec"
	"github.com/beleege/gosrt/util/window"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
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

type ACKAction func(s *SRTSession, seq uint32)

type SRTSession struct {
	conn     net.PacketConn
	peer     net.Addr
	OpenTime time.Time
	RecWin   *window.Entity
	ActList  *list.List
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
	Status   atomic.Value
}

func (s *SRTSession) Write(b []byte) (n int, err error) {
	return s.conn.WriteTo(b, s.peer)
}

func NewSRTSession(c net.PacketConn, a net.Addr) *SRTSession {
	s := new(SRTSession)
	s.conn = c
	s.peer = a
	s.OpenTime = time.Now()
	s.RecWin = window.New(1024, func(seq uint32) {
		if s.ActList.Len() > 0 {
			e := s.ActList.Front()
			if f, ok := e.Value.(ACKAction); ok {
				f(s, seq)
			}
			// clear action list
			s.ActList.Init()
		}
	})
	s.ActList = list.New()
	s.ThisSID = rand.New(rand.NewSource(s.OpenTime.UnixNano())).Uint32()
	s.Status.Store(SNew)
	return s
}

func (s *SRTSession) AddACKAction(f ACKAction) {
	if f != nil {
		s.ActList.PushBack(f)
	}
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
			s.Status.Store(SOpen)
		}
	} else if cif.Version == srt.HSv5 {
		if cif.Cookie != s.Cookie {
			log.Errorf("cookie[%d] is not match", cif.Cookie)
			s.Status.Store(SIllegal)
			return
		}
		if cif.HType == srt.HSTypeConclusion {
			s.Status.Store(SRepeat)
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
			s.TSBPD = srt.ParseHExtension(ext.EContent)
		case srt.HSExtTypeSID:
			s.StreamID = string(ext.EContent)
		}
	}
}

func parseMultiExt(b []byte) []*srt.HSExtension {
	exts := make([]*srt.HSExtension, 0, 2)
	for len(b) > 0 {
		hse := new(srt.HSExtension)

		b = codec.Decode16u(b, &hse.EType)
		b = codec.Decode16u(b, &hse.ELength)
		l := int(hse.ELength) * 4
		hse.EContent = b[:l]
		b = b[l:]

		exts = append(exts, hse)
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
