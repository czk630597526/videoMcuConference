package conference

import (
	"McuConference/libvideo"
	"context"
	"fmt"
	"sync"
	"time"
)

type ConferenceNode struct {
	Rate         int
	Bitrate      int
	mLock        sync.Mutex
	Ctx          context.Context
	cancel       context.CancelFunc
	Name         string
	Pwd          string
	lvcMix       *libvideo.LVC_VideoMixST
	WidTh        int                          //宽
	Height       int                          //高
	lvcEnc       *libvideo.LVC_VideoEncoderST //编码器
	MemberArray  []*MemberNode
	OutBuffer    *AVRing
	eventQueue   chan interface{}
	ts           int
	baseTs       uint32
	ConfIndex    int
	dispatchTime int
	wg           sync.WaitGroup
}

const (
	Low    = "320p"
	Normal = "720P"
	High   = "1080p"
)

func init() {
	libvideo.LVC_Init(4, "hello.log", true)
}

func CreateConference(usr string, pwd string, hdr string, rate, bitrate int) *ConferenceNode {
	width := 720
	height := 480
	if hdr == Low {
		width = 480
		height = 320
	} else if hdr == Normal {
		width = 1280
		height = 720
	} else if hdr == High {
		width = 1980
		height = 1080
	}

	node := &ConferenceNode{
		Ctx:          nil,
		cancel:       nil,
		Name:         usr,
		Pwd:          pwd,
		lvcMix:       nil,
		WidTh:        width,
		Height:       height,
		Rate:         rate,
		Bitrate:      bitrate,
		MemberArray:  make([]*MemberNode, 0),
		eventQueue:   make(chan interface{}),
		lvcEnc:       nil,
		ts:           90000 / rate,
		dispatchTime: 1000 / rate,
	}

	return node
}

func (c *ConferenceNode) StartConference() error {
	c.mLock.Lock()
	defer c.mLock.Unlock()
	if c.Ctx != nil {
		return fmt.Errorf("has started")
	}
	c.Ctx, c.cancel = context.WithCancel(context.Background())
	c.OutBuffer = &AVRing{
		Ring:    nil,
		Context: nil,
		Size:    0,
		poll:    0,
		Type:    0,
	}

	c.OutBuffer.Init(c.Ctx, 20, 1)

	//0表示用openh264编码
	c.lvcEnc = libvideo.LVC_CreatEncoder(0, c.WidTh, c.Height, c.Rate, c.Bitrate, 0)
	if c.lvcEnc == nil {
		return fmt.Errorf("creat enc error")
	}
	go c.runLoopEventQueue()
	go c.runLoopConference()
	return nil
}

func (c *ConferenceNode) MainHandle() {
	c.mLock.Lock()
	defer c.mLock.Unlock()
	if len(c.MemberArray) == 0 {
		return
	}

	numMix := 0
	for _, v := range c.MemberArray {
		if !v.IsNeedIn {
			continue
		}
		if v.subInRing == nil {
			continue
		}
		av, _ := v.subInRing.TryRead()
		if av == nil {
			continue
		}
		//fmt.Printf("read member:%d\n", av.Index)

		dataBuf, ok := av.Value.(*libvideo.LvcFramePicDataSt)
		if !ok {
			continue
		}
		if dataBuf == nil {
			continue
		}
		libvideo.LVC_AddMixPicture(c.lvcMix, dataBuf, v.index)
		v.subInRing.Ring = v.subInRing.Ring.Next()
		numMix++
	}
	if numMix <= 0 {

		return
	}

	out := libvideo.LVC_ProcMixHandle(c.lvcMix)
	if out != nil {
		yuv, ok := libvideo.LVC_VideoDecoderProcFrameLine(out)
		if ok {
			//这里要不编码一次算球了
			nalPkt, ok := libvideo.LVC_VideoEncoderProc(c.lvcEnc, c.ts, yuv)
			if ok {
				//c.baseTs += uint32(c.ts)
				buf := c.OutBuffer.Current()
				buf.Value = nalPkt.Data
				//fmt.Printf("mix %d\n", buf.Index)
				c.OutBuffer.Step()
			}

		}
	}

}

func (c *ConferenceNode) runLoopEventQueue() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case <-c.Ctx.Done():
			{
				//外围cancel
				return
			}
		case <-c.eventQueue:
			{
				//收到消息
			}
		}
	}
}

func (c *ConferenceNode) runLoopConference() {
	//这里应该看帧数决定
	d := time.Millisecond * time.Duration(c.dispatchTime)
	t := time.NewTimer(d)
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case <-c.Ctx.Done():
			{
				//cancel了，可能需要清空内部所有的memberNode
				return
			}
		case <-t.C:
			{
				//在这里混码
				c.MainHandle()
				t.Reset(d)
			}
		}
	}
}

func (c *ConferenceNode) AddMemberNode(node *MemberNode) error {
	if node == nil {
		return fmt.Errorf("none node")
	}
	if node.Conference != nil {
		//已经加过
		return fmt.Errorf("just in any confer")
	}
	l := c.ConfIndex
	if l >= 25 {
		return fmt.Errorf("it is full")
	}

	c.mLock.Lock()
	defer c.mLock.Unlock()

	for _, v := range c.MemberArray {
		if v.Name == node.Name {
			return fmt.Errorf("dup name")
		}
	}

	node.index = l
	node.ts = c.ts
	node.Conference = c
	if node.IsNeedIn {
		if c.lvcMix == nil {
			c.lvcMix = libvideo.LVC_CreateMixHandle(l+1, c.WidTh, c.Height)
		} else {
			libvideo.LVC_DeleteMixHandle(c.lvcMix)
			c.lvcMix = libvideo.LVC_CreateMixHandle(l+1, c.WidTh, c.Height)
		}
		c.ConfIndex++
	} else {
		node.index = -1
	}
	c.MemberArray = append(c.MemberArray, node)
	return nil
}

func (c *ConferenceNode) DelMemberNode(name string) error {
	if name == "" {
		return fmt.Errorf("name param is NULL")
	}
	if len(c.MemberArray) == 0 {
		return fmt.Errorf("usernode is empty")
	}

	c.mLock.Lock()
	defer c.mLock.Unlock()

	var i int
	i = -1

	var lastConfIndex int
	lastConfIndex = -1
	for index, v := range c.MemberArray {
		if v.Name == name {
			i = index
		}
		if v.IsNeedIn && (v.index == c.ConfIndex-1) {
			lastConfIndex = index
		}
	}
	if i == -1 {
		return fmt.Errorf("usernode is NULL")
	}

	n := c.MemberArray[i]
	//关闭节点
	n.StopMember()
	if !n.IsNeedIn {
		c.MemberArray = append(c.MemberArray[:i], c.MemberArray[i+1:]...)
		return nil
	}

	if lastConfIndex == i {

	} else if lastConfIndex == -1 {
		//应该要永远进不来
		return fmt.Errorf("find error")
	} else {
		lastNode := c.MemberArray[lastConfIndex]
		lastNode.index = n.index
	}
	c.MemberArray = append(c.MemberArray[:i], c.MemberArray[i+1:]...)
	c.ConfIndex--

	libvideo.LVC_DeleteMixHandle(c.lvcMix)
	c.lvcMix = libvideo.LVC_CreateMixHandle(c.ConfIndex, c.WidTh, c.Height)

	return nil
}

func (c *ConferenceNode) UpdateNodeIndex(name string, newIndex int) error {
	if name == "" {
		return fmt.Errorf("name param is NULL")
	}
	if len(c.MemberArray) == 0 {
		return fmt.Errorf("usernode is empty")
	}

	c.mLock.Lock()
	defer c.mLock.Unlock()
	var i int
	i = -1

	var ConfIndex int
	ConfIndex = -1
	for index, v := range c.MemberArray {
		if v.Name == name {
			i = index
		}
		if v.IsNeedIn && v.index == newIndex {
			ConfIndex = index
		}
	}
	if i == -1 {
		return fmt.Errorf("usernode is NULL")
	}
	if ConfIndex == -1 {
		return fmt.Errorf("find error")
	}
	n1 := c.MemberArray[i]
	n2 := c.MemberArray[ConfIndex]
	tmp := n1.index
	n1.index = n2.index
	n2.index = tmp
	return nil
}

func (c *ConferenceNode) StopConference() {
	c.mLock.Lock()
	defer c.mLock.Unlock()
	if c.cancel != nil {
		c.cancel()
		c.wg.Wait()
	} else {
		return
	}
	c.Ctx = nil
	c.cancel = nil
	if len(c.MemberArray) == 0 {
		return
	}
	//将member里的数据清除
	for _, v := range c.MemberArray {
		v.StopMember()
	}
	if c.lvcMix != nil {
		libvideo.LVC_DeleteMixHandle(c.lvcMix)
		c.lvcMix = nil
	}
	if c.lvcEnc != nil {
		libvideo.LVC_FreeEncoder(c.lvcEnc)
		c.lvcEnc = nil
	}
	c.OutBuffer.StopRing()
}
