package libvideo_codec

import (
	"encoding/binary"
	"github.com/pion/rtp"
)

type RTP2H264 struct {
	first    bool
	timestap uint32
	rtpBuf   []*rtp.Packet
}

type H264Nal struct {
	Timestap uint32
	Buf      []byte
	EndFlag  int
}

func (self *RTP2H264) Decode(rtpPack *rtp.Packet) (nals []*H264Nal) {

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

func (self *RTP2H264) decode(rtpPacks []*rtp.Packet) (nals []*H264Nal) {
	//IsUseSortPack := true
	//
	//sort.SliceStable(rtpPackets, func(i, j int) bool {
	//	if rtpPackets[i].SequenceNumber == 65535 || rtpPackets[j].SequenceNumber == 65535 {
	//		IsUseSortPack = false
	//	}
	//	return rtpPackets[i].SequenceNumber < rtpPackets[j].SequenceNumber
	//})

	var h264nal *H264Nal

	for _, rtpPack := range rtpPacks {

		bufType := int(rtpPack.Payload[0] & 0x1f)

		switch {
		case bufType >= 1 && bufType <= 23:
			if h264nal != nil {
				nals = append(nals, h264nal)
				h264nal = nil
			}

			h264nal = new(H264Nal)
			h264nal.Timestap = rtpPack.Timestamp
			h264nal.Buf = append([]byte{0, 0, 0, 1}, rtpPack.Payload...)

			nals = append(nals, h264nal)
			h264nal = nil

		case bufType == 24: // STAP-A
			if h264nal != nil {
				nals = append(nals, h264nal)
				h264nal = nil
			}

			nalBuf := rtpPack.Payload[1:]
			for len(nalBuf) >= 3 {
				size := int(binary.BigEndian.Uint16(nalBuf))
				if size+2 > len(nalBuf) {
					break
				}

				h264nal = new(H264Nal)
				h264nal.Timestap = rtpPack.Timestamp
				h264nal.Buf = append([]byte{0, 0, 0, 1}, nalBuf[2:size+2]...)

				nals = append(nals, h264nal)
				h264nal = nil

				nalBuf = nalBuf[size+2:]
			}
		case bufType == 28: // FU-A
			if len(rtpPack.Payload) < 3 {
				return
			}

			fuIndicator := rtpPack.Payload[0]
			fuHeader := rtpPack.Payload[1]
			isStart := (fuHeader & 0x80) != 0
			isEnd := (fuHeader & 0x40) != 0
			nalheader := (fuIndicator & 0xe0) | (fuHeader & 0x1f)

			if isStart { //开始包
				if h264nal != nil {
					nals = append(nals, h264nal)
					h264nal = nil
				}
				h264nal = new(H264Nal)
				h264nal.Timestap = rtpPack.Timestamp
				h264nal.Buf = append([]byte{0, 0, 0, 1, nalheader}, rtpPack.Payload[2:]...)
				continue
			}

			if h264nal != nil {
				h264nal.Buf = append(h264nal.Buf, rtpPack.Payload[2:]...)
			} else {
				h264nal = new(H264Nal)
				h264nal.Timestap = rtpPack.Timestamp
				h264nal.Buf = append([]byte{0, 0, 0, 1, nalheader}, rtpPack.Payload[2:]...)
			}

			if isEnd {
				nals = append(nals, h264nal)
				h264nal = nil
			}
		}
	}

	nalsLen := len(nals)
	if nalsLen > 0 {
		nals[nalsLen-1].EndFlag = 1
	}

	return
}
