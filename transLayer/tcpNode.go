package transLayer

import (
	"McuConference/conference"
	"McuConference/libvideo_codec"
	"McuConference/udpmedia"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"io"
	"net"
	"sync"
)

type TcpNodeInfo struct {
	MediaCtx    context.Context
	mediaCancel context.CancelFunc

	Ctx             context.Context
	cancel          context.CancelFunc
	LocalVideoConn  *udpmedia.UdpRTPConn
	RemoteVideoAddr *net.UDPAddr
	roomName        string
	nodeName        string
	conn            net.Conn
	cfrMgr          *conference.ConferenceMgr
	packtizer       rtp.Packetizer
	parse           libvideo_codec.RTP2H264
	mediaChan       chan *rtp.Packet
	wg              sync.WaitGroup
}

var (
	MAX_RTP_PACKET = 128
)

func CreateTcpNode(conn net.Conn) *TcpNodeInfo {
	return &TcpNodeInfo{
		LocalVideoConn:  nil,
		RemoteVideoAddr: nil,
		conn:            conn,
		packtizer:       rtp.NewPacketizer(1200, 96, uint32(54321), &codecs.H264Payloader{}, rtp.NewFixedSequencer(1), 90000),
	}
}

func (node *TcpNodeInfo) Start() error {
	if node.Ctx != nil {
		return fmt.Errorf("has started")
	}
	node.Ctx, node.cancel = context.WithCancel(context.Background())
	go node.runLoop()
	return nil
}

func (node *TcpNodeInfo) Stop() error {
	if node.Ctx == nil {
		return fmt.Errorf("not start")
	}
	if node.cancel != nil {
		node.cancel()
	}
	node.cancel = nil
	node.Ctx = nil
	return nil
}

func (node *TcpNodeInfo) processMsg(buffer []byte) error {
	req := conference.ReqInfo{}
	err := json.Unmarshal(buffer, &req)
	if err != nil {
		return err
	}
	if req.Method == "join" {
		join, ok := req.Param.(*conference.JoinInfo)
		if ok {
			node.handleJoinInfo(join)
		} else {
			return fmt.Errorf("join vaild:%s\n", req.Method)
		}
	} else if req.Method == "leave" {
		join, ok := req.Param.(*conference.LeaveInfo)
		if ok {
			node.handleLeaveInfo(join)
		} else {
			return fmt.Errorf("join vaild:%s\n", req.Method)
		}
	} else if req.Method == "update" {
		join, ok := req.Param.(*conference.UpdateInfo)
		if ok {
			node.handleUpdateInfo(join)
		} else {
			return fmt.Errorf("join vaild:%s\n", req.Method)
		}
	} else {
		return fmt.Errorf("method vaild:%s\n", req.Method)
	}
	return nil
}

func (node *TcpNodeInfo) handleJoinInfo(info *conference.JoinInfo) error {
	if info == nil {
		return fmt.Errorf("no param")
	}

	if node.roomName != "" || node.nodeName != "" {
		b := MakeRsp(&conference.RspInfo{
			Method:   "join",
			ErrCode:  104,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   "client already join",
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("client already join")
	}

	addr, err := net.ResolveUDPAddr("udp", info.RemoteVAddr)
	if err != nil {
		b := MakeRsp(&conference.RspInfo{
			Method:   "join",
			ErrCode:  101,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   "remote addr error",
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("remote addr error")
	}

	udpConn := udpmedia.NewUdpRtpConn()
	if udpConn == nil {
		b := MakeRsp(&conference.RspInfo{
			Method:   "join",
			ErrCode:  102,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   "get rtp error",
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("get rtp error")
	}

	cb := conference.UserNodeCallBak{}
	cb.FuncOnOutput = node.onOutPut
	cb.FuncOnInput = node.OnInput
	cb.FuncOnClose = node.onClose
	cb.FuncOnCloseDelay = node.onCloseDelay

	err = node.cfrMgr.JoinMemberNode(info, cb)
	if err != nil {
		udpConn.Close()
		b := MakeRsp(&conference.RspInfo{
			Method:   "join",
			ErrCode:  103,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   err.Error(),
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("join error")
	}
	b := MakeRsp(&conference.RspInfo{
		Method:   "join",
		ErrCode:  0,
		RoomName: info.RoomName,
		NodeName: info.NodeName,
		ErrMsg:   "Succ",
		Param: &conference.MediaInfo{
			LocalVAddr: udpConn.Rtpaddr.String(),
		},
	})

	node.nodeName = info.NodeName
	node.roomName = info.RoomName
	node.RemoteVideoAddr = addr
	node.LocalVideoConn = udpConn
	node.conn.Write(b)
	go node.handleMedia()
	return nil
}

func (node *TcpNodeInfo) handleMedia() {
	node.MediaCtx, node.mediaCancel = context.WithCancel(context.Background())
	node.mediaChan = make(chan *rtp.Packet, MAX_RTP_PACKET)
	node.wg.Add(1)
	defer node.wg.Done()
	defer close(node.mediaChan)
	node.mediaChan = nil
	for {
		select {
		case <-node.MediaCtx.Done():
			return
		default:
			r, _, e := node.LocalVideoConn.ReadRtp()
			if e != nil {
				continue
			}
			if len(node.mediaChan) >= MAX_RTP_PACKET {
				continue
			}
			node.mediaChan <- r
		}
	}
}

func (node *TcpNodeInfo) handleLeaveInfo(info *conference.LeaveInfo) error {
	if info == nil {
		return fmt.Errorf("no param")
	}
	if info.RoomName != node.roomName || info.NodeName != node.nodeName {
		b := MakeRsp(&conference.RspInfo{
			Method:   "leave",
			ErrCode:  101,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   "not join",
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("no join")
	}

	err := node.cfrMgr.LeaveMemberNode(info)
	if err != nil {
		b := MakeRsp(&conference.RspInfo{
			Method:   "leave",
			ErrCode:  102,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   err.Error(),
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("no join")
	}

	b := MakeRsp(&conference.RspInfo{
		Method:   "leave",
		ErrCode:  0,
		RoomName: info.RoomName,
		NodeName: info.NodeName,
		ErrMsg:   "Succ",
		Param:    nil,
	})
	node.conn.Write(b)
	return nil
}

func (node *TcpNodeInfo) handleUpdateInfo(info *conference.UpdateInfo) error {
	if info == nil {
		return fmt.Errorf("no param")
	}
	if info.RoomName != node.roomName || info.NodeName != node.nodeName {
		b := MakeRsp(&conference.RspInfo{
			Method:   "update",
			ErrCode:  101,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   "not join",
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("no join")
	}
	err := node.cfrMgr.UpdateMemberNode(info)
	if err != nil {
		b := MakeRsp(&conference.RspInfo{
			Method:   "update",
			ErrCode:  102,
			RoomName: info.RoomName,
			NodeName: info.NodeName,
			ErrMsg:   err.Error(),
			Param:    nil,
		})
		node.conn.Write(b)
		return fmt.Errorf("update error")
	}
	b := MakeRsp(&conference.RspInfo{
		Method:   "update",
		ErrCode:  0,
		RoomName: info.RoomName,
		NodeName: info.NodeName,
		ErrMsg:   "Succ",
		Param:    nil,
	})
	node.conn.Write(b)

	return nil
}

func (node *TcpNodeInfo) runLoop() {
	defer func() {
		if node.nodeName != "" && node.roomName != "" {
			node.cfrMgr.LeaveMemberNode(&conference.LeaveInfo{
				RoomName: node.roomName,
				NodeName: node.nodeName,
			})
		}
		node.conn.Close()
	}()
	for {
		select {
		case <-node.Ctx.Done():
			return
		default:
			head := make([]byte, 6)
			_, err := io.ReadFull(node.conn, head)
			if nil != err {
				fmt.Printf("io.ReadFull() failed, error[%s]", err.Error())
				return
			}
			if head[0] == 'w' && head[1] == 'e' && head[2] == 'r' && head[3] == 'x' {
				bodyLen := binary.BigEndian.Uint16(head[4:])
				buffer := make([]byte, bodyLen)
				_, err := io.ReadFull(node.conn, buffer)
				if nil != err {
					fmt.Printf("io.ReadFull() buffer failed error[%s]", err.Error())
					return
				}
				err = node.processMsg(buffer)
				if nil != err {
					fmt.Printf("processMsg failed error[%s]", err.Error())
					return
				}
			} else {
				fmt.Printf("io.ReadFull() buffer check failed")
				return
			}
		}
	}
}
