package demoConfMgr

import (
	"McuConference/conference"
	"McuConference/libvideo_codec"
	"fmt"
	"github.com/pion/rtp"
	"os"
	"sync"
	"time"
)

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
