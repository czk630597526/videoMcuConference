package transLayer

import (
	"McuConference/libvideo_codec"
)

func (node *TcpNodeInfo) OnInput() []*libvideo_codec.H264Nal {
	r := <-node.mediaChan
	if r == nil {
		return nil
	}
	return node.parse.Decode(r)
}

func (node *TcpNodeInfo) onClose() {
	if node.mediaCancel == nil {
		return
	}
	node.roomName = ""
	node.nodeName = ""
	node.mediaCancel()
	node.mediaCancel = nil
	node.wg.Wait()
}

func (node *TcpNodeInfo) onCloseDelay() {
	if node.LocalVideoConn != nil {
		node.LocalVideoConn.Close()
		node.LocalVideoConn = nil
	}
}

func (node *TcpNodeInfo) onOutPut(nal []byte, ts uint32) {
	packets := node.packtizer.Packetize(nal, ts)
	for packIndex, packet := range packets {
		packet.Marker = packIndex == len(packets)-1
		rtpBuf, _ := packet.Marshal()
		node.LocalVideoConn.Write(rtpBuf, node.RemoteVideoAddr)
	}
}
