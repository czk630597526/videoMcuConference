package udpmedia

import (
	"fmt"
	"github.com/pion/rtp"
	"github.com/pkg/errors"
	"net"
	"time"
)

type UdpRTPConn struct {
	Rtpconn  *net.UDPConn
	Rtpaddr  *net.UDPAddr
	Rtpport  int
	Rtcpconn *net.UDPConn
	Rtcpaddr *net.UDPAddr
	Rtcpport int
	ServerIp string
}

var udpStartPort = 0
var udpEndPort = 0
var nowUdpPort = 0
var serverIp = ""

func MediaInit(startUdpPort, endUdpPort int, ip string) {
	udpStartPort = startUdpPort
	udpEndPort = endUdpPort
	nowUdpPort = startUdpPort
	serverIp = ip
}

func NewUdpRtpConn() (udpConn *UdpRTPConn) {
	var err error
	if nowUdpPort > udpEndPort {
		nowUdpPort = udpStartPort
	}

	udpConn = &UdpRTPConn{ServerIp: serverIp}
	for i := nowUdpPort; i < udpEndPort; i = i + 2 {
		udpConn.Rtpaddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIp, i))
		if err != nil {
			continue
		}
		udpConn.Rtcpaddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIp, i+1))
		if err != nil {
			continue
		}
		if udpConn.Rtpconn, err = net.ListenUDP("udp", udpConn.Rtpaddr); err != nil {
			continue
		} else if udpConn.Rtcpconn, err = net.ListenUDP("udp", udpConn.Rtcpaddr); err != nil {
			_ = udpConn.Rtpconn.Close()
			udpConn.Rtpconn = nil
			udpConn.Rtcpconn = nil
		}
		udpConn.Rtpport = i
		udpConn.Rtcpport = i + 1
		nowUdpPort = i + 2
		return
	}

	for i := udpStartPort; i < nowUdpPort; i = i + 2 {
		udpConn.Rtpaddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIp, i))
		if err != nil {
			continue
		}
		udpConn.Rtcpaddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIp, i+1))
		if err != nil {
			continue
		}
		if udpConn.Rtpconn, err = net.ListenUDP("udp", udpConn.Rtpaddr); err != nil {
			continue
		} else if udpConn.Rtcpconn, err = net.ListenUDP("udp", udpConn.Rtcpaddr); err != nil {
			_ = udpConn.Rtpconn.Close()
			udpConn.Rtpconn = nil
			udpConn.Rtcpconn = nil
		}
		udpConn.Rtpport = i
		udpConn.Rtcpport = i + 1
		nowUdpPort = i + 2
		return
	}

	return nil
}

func (self *UdpRTPConn) Close() {
	if self.Rtpconn != nil {
		_ = self.Rtpconn.Close()
		self.Rtpconn = nil
		self.Rtpaddr = nil
	}
	if self.Rtcpconn != nil {
		_ = self.Rtcpconn.Close()
		self.Rtcpconn = nil
		self.Rtcpaddr = nil
	}
}

func (self *UdpRTPConn) SetReadTimeout(duration time.Duration) {
	self.Rtpconn.SetReadDeadline(time.Now().Add(duration))
}

func (self *UdpRTPConn) CancelReadTimeout() {
	self.Rtpconn.SetReadDeadline(time.Time{})
}

func (self *UdpRTPConn) Read(b []byte) (int, *net.UDPAddr, error) {
	if self.Rtpconn == nil {
		return 0, nil, errors.New("rtp conn is nil")
	}
	return self.Rtpconn.ReadFromUDP(b)
}

func (self *UdpRTPConn) ReadRtp() (*rtp.Packet, *net.UDPAddr, error) {
	b := make([]byte, 1500)
	n, addr, err := self.Rtpconn.ReadFromUDP(b)
	if err != nil {
		return nil, nil, err
	}

	r := &rtp.Packet{}
	if err := r.Unmarshal(b[:n]); err != nil {
		return nil, nil, err
	}
	return r, addr, nil
}

func (self *UdpRTPConn) ReadRtcp(b []byte) (int, *net.UDPAddr, error) {
	if self.Rtpconn == nil {
		return 0, nil, errors.New("rtcp conn is nil")
	}
	return self.Rtcpconn.ReadFromUDP(b)
}

func (self *UdpRTPConn) Write(b []byte, addr *net.UDPAddr) (int, error) {
	if self.Rtpconn == nil {
		return 0, errors.New("rtp conn is nil")
	}

	return self.Rtpconn.WriteToUDP(b, addr)
}
