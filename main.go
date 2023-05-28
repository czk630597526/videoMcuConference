package main

import "C"
import (
	"McuConference/conference"
	"McuConference/libvideo"
	"McuConference/libvideo_codec"
	"fmt"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"os"
	"sync"
	"time"
)

func mainDecode() {
	libvideo.LVC_Init(4, "hello.log", false)
	lvcDec := libvideo.LVC_CreatDecoder(1)
	if lvcDec == nil {
		fmt.Printf("get decoder error\n")
		return
	}
	f, _ := os.Open("./test.h264")
	d, _ := os.Create("./out.yuv")
	b := make([]byte, 1024*1024*5)
	l, _ := f.Read(b)
	start := 0
	pos := 0
	flag := 0
	buffer := b[:l]
	fmt.Printf("get buffer:%d\n", l)
	for {
		if start >= l {
			break
		}
		if start+4 < l && buffer[start] == 0 && buffer[start+1] == 0 && buffer[start+2] == 0 && buffer[start+3] == 1 {
			if flag == 0 {
				pos = start
				start++
				flag = 1
				continue
			}
			out := buffer[pos:start]
			yuv, ok := libvideo.LVC_VideoDecoderProcSimple2(lvcDec, out)
			if ok {
				//for _, v := range yuv
				{
					fmt.Printf("write:%d ,%d,%d\n", pos, start, len(yuv))
					d.Write(yuv)
				}
			}

			fmt.Printf("write:%d ,%d\n", pos, start)
			pos = start
		}
		start++
	}

}

func mainMix() {
	libvideo.LVC_Init(4, "hello.log", false)
	lvcDec := libvideo.LVC_CreatDecoder(1)
	if lvcDec == nil {
		fmt.Printf("get decoder error\n")
		return
	}
	f, _ := os.Open("./test.h264")
	d, _ := os.Create("./out.h264")
	b := make([]byte, 1024*1024*5)
	l, _ := f.Read(b)
	start := 0
	pos := 0
	flag := 0
	buffer := b[:l]
	num := 1
	fmt.Printf("get buffer:%d\n", l)
	lvcMix := libvideo.LVC_CreateMixHandle(num, 480, 320)
	if lvcMix == nil {
		fmt.Printf("create mix error\n")
		return
	}
	pdata := libvideo.LvcFramePicData_Create(0, 0, 0)
	lvcEnc := libvideo.LVC_CreatEncoder(0, 480, 320, 15, 1024, 0)
	if lvcEnc == nil {
		fmt.Printf("create mix error\n")
		return
	}
	i := 0
	j := 0
	ts := 0
	tsInc := 6000
	for {
		if start >= l {
			break
		}
		if start+4 < l && buffer[start] == 0 && buffer[start+1] == 0 && buffer[start+2] == 0 && buffer[start+3] == 1 {
			if flag == 0 {
				pos = start
				start++
				flag = 1
				continue
			}
			out := buffer[pos:start]

			ok := libvideo.LVC_VideoDecoderProcFrameNoLine(lvcDec, pdata, out)
			if ok {
				for i = 0; i < num; i++ {
					libvideo.LVC_AddMixPicture(lvcMix, pdata, i)
				}
				pMix := libvideo.LVC_ProcMixHandle(lvcMix)
				if pMix != nil {
					yuv, ok := libvideo.LVC_VideoDecoderProcFrameLine(pMix)
					if ok {
						j++
						fmt.Printf("write:%d ,%d,%d,%d\n", pos, start, len(yuv), i)
						nalPkt, ok := libvideo.LVC_VideoEncoderProc(lvcEnc, ts, yuv)

						ts = ts + tsInc
						if ok {
							d.Write(nalPkt.Data)
						}

						//d.Write(yuv)
					}
				}
			}
			pos = start
		}
		start++
	}
	libvideo.LvcFramePicData_Delete(pdata)
	libvideo.LVC_DeleteMixHandle(lvcMix)
}

func TransRtp2H264(b []byte, ssrc []byte) (nal [][]byte) {
	if len(ssrc) != 4 {
		return nil
	}
	var index int
	var flg bool
	index = 0
	var start int
	start = 0
	var end int
	end = 0

	for i, v := range b {
		if v == ssrc[index] {
			index++
		} else {
			index = 0
		}

		if index == 4 {
			index = 0
			end = i
			if flg == false {
				start = i
				flg = true
			} else {
				fmt.Printf("get start:%d,%d\n", start-11, end-11)
				rawRtp := b[start-11 : end-11]
				start = i
				nal = append(nal, rawRtp)
				//stopNum++
				//if stopNum == 2 {
				//	return nal
				//}
			}
		}
	}
	return
}

func mainRtp() {
	libvideo.LVC_Init(4, "hello.log", false)
	lvcDec := libvideo.LVC_CreatDecoder(1)
	if lvcDec == nil {
		fmt.Printf("get decoder error\n")
		return
	}
	f, e := os.OpenFile("Newvideo.rtp", os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Printf("read rtp error")
	}
	lvcEnc := libvideo.LVC_CreatEncoder(0, 480, 640, 25, 1024, 0)
	if lvcEnc == nil {
		fmt.Printf("create mix error\n")
		return
	}
	buf := make([]byte, 1024*1024*3)
	l, e := f.Read(buf)
	b := buf[:l]
	n := TransRtp2H264(b, []byte{0xE6, 0x5F, 0x87, 0x19})

	f1, _ := os.Create("video3.h264")
	f2, _ := os.Create("video4.rtp")
	f3, _ := os.Create("video5.h264")
	parse := libvideo_codec.RTP2H264{}
	pdata := libvideo.LvcFramePicData_Create(0, 0, 0)
	baseTs := 0
	ts := 4000
	i := 0
	vpacketer := rtp.NewPacketizer(1200, 96, uint32(54321), &codecs.H264Payloader{}, rtp.NewFixedSequencer(1), 90000)
	for _, v := range n {
		p := &rtp.Packet{}
		p.Unmarshal(v)
		nals := parse.Decode(p)
		for _, nal := range nals {
			f3.Write(nal.Buf)
			ok := libvideo.LVC_VideoDecoderProcFrameNoLine(lvcDec, pdata, nal.Buf)
			if ok {
				yuv, ok := libvideo.LVC_VideoDecoderProcFrameLine(pdata)
				if ok {
					i++
					fmt.Printf("yuv:%d\n", i)
					nalPkt, ok := libvideo.LVC_VideoEncoderProc(lvcEnc, ts, yuv)
					baseTs += ts
					if ok {
						f1.Write(nalPkt.Data)
						packets := vpacketer.Packetize(nalPkt.Data, uint32(ts))
						for packIndex, packet := range packets {
							packet.Marker = packIndex == len(packets)-1
							rtpBuf, _ := packet.Marshal()
							fmt.Printf("write:%d,%d\n", packet.Timestamp, packet.SequenceNumber)
							f2.Write(rtpBuf)
						}
					}

					//d.Write(yuv)
				}
			}
		}
		//time.Sleep(time.Millisecond * 30)
	}
	libvideo.LvcFramePicData_Delete(pdata)
}

func MainConference() {
	nodeArray := make([]*conference.MemberNode, 0)

	room := conference.CreateConference("foo", "bar", conference.Normal, 25, 1024)
	if room == nil {
		panic("room is nil")
	}
	room.StartConference()
	f3, _ := os.Create("mix.h264")
	for i := 0; i < 10; i++ {
		node := conference.CreateMember(fmt.Sprintf("conf:%d", i))
		if node == nil {
			panic("node is nil")
		}
		node.SetInBuffer(true)
		node.SetOutBuffer(false)
		if i == 0 {
			node.SetOutBuffer(true)
			node.SetOnOutPut(func(bytes []byte, u uint32) {
				f3.Write(bytes)
			})
		}
		room.AddMemberNode(node)
		node.StartMember()
		nodeArray = append(nodeArray, node)
	}

	f, _ := os.Open("./video2.h264")
	b := make([]byte, 1024*1024*5)
	l, _ := f.Read(b)
	start := 0
	pos := 0
	flag := 0
	buffer := b[:l]

	allLen := len(nodeArray) - 1

	incFlg := false

	t := time.NewTimer(2 * time.Second)
	for {

		select {
		case <-t.C:
			if incFlg == false {
				if allLen == 0 {
					t.Stop()
					incFlg = true
					break
				}
				room.DelMemberNode(fmt.Sprintf("conf:%d", allLen))
				nodeArray = append(nodeArray[:allLen], nodeArray[allLen+1:]...)
				allLen--
			} else {
				allLen++
				if allLen >= 10 {
					t.Stop()
					break
				}

				node := conference.CreateMember(fmt.Sprintf("conf:%d", allLen))
				if node == nil {
					panic("node is nil")
				}
				node.SetInBuffer(true)
				node.SetOutBuffer(false)
			}

			t.Reset(2 * time.Second)

		default:
			if start >= l {
				break
			}
			if start+4 < l && buffer[start] == 0 && buffer[start+1] == 0 && buffer[start+2] == 0 && buffer[start+3] == 1 {
				if flag == 0 {
					pos = start
					start++
					flag = 1
					continue
				}
				fmt.Printf("write:%d,%d,%d\n", pos, start, l)
				out := buffer[pos:start]
				for _, node := range nodeArray {
					node.WriteStream(out)
				}
				time.Sleep(time.Millisecond * 40)
				pos = start
			}
			start++
		}
	}

	//select {}
}

type nodeTmp struct {
	name      string
	n         [][]byte
	mediaChan chan []*libvideo_codec.H264Nal
	parse     libvideo_codec.RTP2H264
}

func (n *nodeTmp) start() {
	for _, v := range n.n {
		p := &rtp.Packet{}
		p.Unmarshal(v)
		nals := n.parse.Decode(p)
		if nals == nil {
			continue
		}
		n.mediaChan <- nals
		//fmt.Printf("in buf\n")
		time.Sleep(time.Millisecond * 30)
	}
}

func (n *nodeTmp) onInput() []*libvideo_codec.H264Nal {
	//fmt.Printf("get buf\n")
	b := <-n.mediaChan
	return b
}

func (n *nodeTmp) onClose() {
	close(n.mediaChan)
	fmt.Printf("on close:%s\n", n.name)
}

func (n *nodeTmp) onCloseDelay() {
	fmt.Printf("on delayclose:%s\n", n.name)
}

func MainConferenceMgr() {
	mgr := conference.CreateCfrMgr()
	mgr.AddConferenceNode(&conference.CreateConfInfo{
		Name:      "foo",
		Pwd:       "bar",
		HdrString: conference.Normal,
		Rate:      25,
		Bitrate:   1024,
	})

	f3, _ := os.Create("mix.h264")
	buf := make([]byte, 1024*1024*3)
	f, e := os.OpenFile("Newvideo.rtp", os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Printf("read rtp error")
	}
	l, e := f.Read(buf)
	b := buf[:l]
	n := TransRtp2H264(b, []byte{0xE6, 0x5F, 0x87, 0x19})

	wg := sync.WaitGroup{}

	mgr.JoinMemberNode(&conference.JoinInfo{
		RoomName:    "foo",
		NodeName:    fmt.Sprintf("conf:%d", 20),
		IsIn:        false,
		IsOut:       true,
		RemoteVAddr: "",
	}, conference.UserNodeCallBak{
		FuncOnOutput: func(bytes []byte, u uint32) {
			f3.Write(bytes)
		},
	})

	for i := 0; i < 10; i++ {
		node := nodeTmp{
			name:      fmt.Sprintf("conf:%d", i),
			n:         n,
			parse:     libvideo_codec.RTP2H264{},
			mediaChan: make(chan []*libvideo_codec.H264Nal, 128),
		}

		info := &conference.JoinInfo{
			RoomName:    "foo",
			NodeName:    fmt.Sprintf("conf:%d", i),
			IsIn:        true,
			IsOut:       false,
			RemoteVAddr: "",
		}
		cb := conference.UserNodeCallBak{
			FuncOnCloseDelay: node.onCloseDelay,
			FuncOnClose:      node.onClose,
			FuncOnInput:      node.onInput,
			FuncOnOutput:     nil,
		}

		mgr.JoinMemberNode(info, cb)

		go func(rn string, nn string) {
			wg.Add(1)
			fmt.Printf("start inner %s\n", rn)
			defer wg.Done()
			node.start()
			mgr.LeaveMemberNode(&conference.LeaveInfo{
				RoomName: rn,
				NodeName: nn,
			})
			fmt.Printf("stop inner %s\n", rn)
		}(info.RoomName, info.NodeName)
		time.Sleep(time.Second)
	}
	wg.Wait()
	//select {}
}

func TestReadTmp() {
	buf := make([]byte, 1024*1024*3)
	f, e := os.OpenFile("Newvideo.rtp", os.O_RDWR|os.O_CREATE, 0755)
	if e != nil {
		fmt.Printf("read rtp error")
	}
	l, e := f.Read(buf)
	b := buf[:l]
	n := TransRtp2H264(b, []byte{0xE6, 0x5F, 0x87, 0x19})
	parse := libvideo_codec.RTP2H264{}
	fmt.Printf("%s\n", time.Now().String())
	for _, v := range n {
		p := &rtp.Packet{}
		p.Unmarshal(v)
		nals := parse.Decode(p)
		if nals == nil {
			continue
		}
		//fmt.Printf("in buf\n")
		time.Sleep(time.Millisecond * 30)
	}
	fmt.Printf("%s\n", time.Now().String())
}

func main() {
	//mainDecode()
	//mainRtp()
	//mainMix()
	MainConferenceMgr()
	//MainConference()
	//TestReadTmp()
	//fmt.Printf("%s\n", time.Now().String())
	//for i := 0; i < 100; i++ {
	//	time.Sleep(time.Millisecond * 30)
	//}
	//fmt.Printf("%s\n", time.Now().String())
}
