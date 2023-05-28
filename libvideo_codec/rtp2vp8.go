package libvideo_codec

import (
	"github.com/pion/rtp"
)

type RTP2VP8 struct {
	first    bool
	timestap uint32
	rtpBuf   []*rtp.Packet
}

type VP8Nal struct {
	Timestap uint32
	Buf      []byte
	EndFlag  int
}

func (self *RTP2VP8) Decode(rtpPack *rtp.Packet) (nals []*VP8Nal) {

	if !self.first {
		self.first = true
		self.timestap = rtpPack.Timestamp
	}

	if self.timestap == rtpPack.Timestamp {
		self.rtpBuf = append(self.rtpBuf, rtpPack)
		return
	}

	self.timestap = rtpPack.Timestamp
	rtpPackets := self.rtpBuf
	self.rtpBuf = []*rtp.Packet{rtpPack}

	return self.decode(rtpPackets)
}

func (self *RTP2VP8) decode(rtppackets []*rtp.Packet) (nals []*VP8Nal) {
	//IsUseSortPack := true
	//
	//sort.SliceStable(rtpPackets, func(i, j int) bool {
	//	if rtpPackets[i].SequenceNumber == 65535 || rtpPackets[j].SequenceNumber == 65535 {
	//		IsUseSortPack = false
	//	}
	//	return rtpPackets[i].SequenceNumber < rtpPackets[j].SequenceNumber
	//})

	if len(rtppackets) < 1 {
		return
	}

	Frame := new(VP8Nal)
	Frame.Timestap = rtppackets[0].Timestamp
	Frame.Buf = make([]byte, 0)

	for _, rtppacket := range rtppackets {
		vp8descriptor := Vp8Descriptor_Unmarshal(rtppacket.Payload)
		if vp8descriptor == nil || len(vp8descriptor.Payload) < 1 {
			continue
		}

		Frame.Buf = append(Frame.Buf, vp8descriptor.Payload...)
	}

	Frame.EndFlag = 1
	nals = append(nals, Frame)

	return
}

type VP8_DESCRIPTOR struct {
	X   bool
	R   bool
	N   bool
	S   bool
	PID byte

	I bool
	L bool
	T bool
	K bool

	M         bool
	PictureId uint16

	LContent byte

	TKContent byte

	Payload []byte
}

func Vp8Descriptor_Unmarshal(payload []byte) *VP8_DESCRIPTOR {
	if len(payload) < 1 {
		return nil
	}

	descriptor := new(VP8_DESCRIPTOR)
	if descriptor.Unmarshal(payload) {
		return descriptor
	} else {
		return nil
	}
}

func (self *VP8_DESCRIPTOR) Unmarshal(payload []byte) bool {
	if len(payload) < 1 {
		return false
	}

	self.Payload = payload

	self.X = ((self.Payload[0] & 0x80) == 0x80)
	self.R = ((self.Payload[0] & 0x40) == 0x40)
	self.N = ((self.Payload[0] & 0x20) == 0x20)
	self.S = ((self.Payload[0] & 0x10) == 0x10)
	self.PID = (self.Payload[0] & 0x0F)
	self.Payload = self.Payload[1:]

	if self.X {
		if len(self.Payload) < 1 {
			return false
		}
	} else {
		return true
	}

	self.I = ((self.Payload[0] & 0x80) == 0x80)
	self.L = ((self.Payload[0] & 0x40) == 0x40)
	self.T = ((self.Payload[0] & 0x20) == 0x20)
	self.K = ((self.Payload[0] & 0x10) == 0x10)
	self.Payload = self.Payload[1:]

	if self.I {
		if len(self.Payload) < 1 {
			return false
		} else {
			self.M = ((self.Payload[0] & 0x80) == 0x80)
			self.PictureId = uint16(self.Payload[0] & 0x7F)
			self.Payload = self.Payload[1:]

			if self.M {
				if len(self.Payload) < 1 {
					return false
				} else {
					self.PictureId = (self.PictureId << 8) + uint16(self.Payload[0])
					self.Payload = self.Payload[1:]
				}
			}

		}
	}

	if self.L {
		if len(self.Payload) < 1 {
			return false
		} else {
			self.LContent = self.Payload[0]
			self.Payload = self.Payload[1:]
		}
	}

	if self.T || self.K {
		if len(self.Payload) < 1 {
			return false
		} else {
			self.TKContent = self.Payload[0]
			self.Payload = self.Payload[1:]
		}
	}

	return true
}
