package conference

import (
	"McuConference/libvideo"
	"McuConference/libvideo_codec"
	"context"
	"fmt"
	"sync"
	"time"
)

type MemberNode struct {
	memLock     sync.Mutex
	Name        string
	Ctx         context.Context
	cancel      context.CancelFunc
	lvcDec      *libvideo.LVC_VideoDecoderST //解码器
	InBuffer    *AVRing                      //inbuffer，需要将其解码为yuv
	OutBuffer   *AVRing                      //outBUffer，内部是个yuv，是个subring
	IsNeedIn    bool                         //是否需要in
	IsNeedOut   bool                         //是否需要out
	IsNetWork   bool
	index       int //混屏的位置
	ts          int
	streamQueue chan []byte //裸264流
	Conference  *ConferenceNode
	subInRing   *AVRing
	deBug       bool
	onClose     func()

	onOutPut func([]byte, uint32)
	onInPut  func() []*libvideo_codec.H264Nal

	onCLoseDelay func()
	wg           sync.WaitGroup
}

func CreateMember(name string) *MemberNode {
	return &MemberNode{
		Name:        name,
		Ctx:         nil,
		cancel:      nil,
		lvcDec:      nil,
		InBuffer:    nil,
		OutBuffer:   nil,
		IsNeedIn:    false,
		IsNeedOut:   false,
		index:       0,
		streamQueue: make(chan []byte, 64),
	}
}

func (m *MemberNode) SetInBuffer(flg bool) {
	m.IsNeedIn = flg
}

func (m *MemberNode) SetDebug(flg bool) {
	m.deBug = flg
}

func (m *MemberNode) SetOutBuffer(flg bool) {
	m.IsNeedOut = flg
}

func (m *MemberNode) WriteStream(b []byte) {
	//fmt.Printf("start write %s\n", m.Name)
	if m.cancel == nil {
		return
	}
	m.streamQueue <- b
	//fmt.Printf("end write %s\n", m.Name)
}

func (m *MemberNode) SetOnClose(f func()) {
	m.onClose = f
}

func (m *MemberNode) SetOnCloseDelay(f func()) {
	m.onCLoseDelay = f
}

func (m *MemberNode) SetOnOutPut(f func([]byte, uint32)) {
	m.onOutPut = f
}

func (m *MemberNode) SetOnInput(f func() []*libvideo_codec.H264Nal) {
	m.onInPut = f
}

func (m *MemberNode) StartMember() bool {
	m.memLock.Lock()
	defer m.memLock.Unlock()
	if m.Ctx != nil {
		return false
	}
	if m.Conference == nil {
		return false
	}
	m.Ctx, m.cancel = context.WithCancel(context.Background())
	//1表示用openh264解码
	m.lvcDec = libvideo.LVC_CreatDecoder(1)
	if m.lvcDec == nil {
		return false
	}
	if m.IsNeedIn {
		m.InBuffer = &AVRing{
			Type: 0,
			poll: time.Microsecond * 10,
		}
		m.InBuffer.Init(m.Ctx, 20, 0)
	}
	if m.IsNeedOut {

	}
	go m.runLoopIn()
	go m.runLoopDec()
	go m.runLoopOut()
	return true
}

func (m *MemberNode) runLoopIn() {
	if !m.IsNeedIn {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	for {
		select {
		case <-m.Ctx.Done():
			{
				//调用cancel了
				fmt.Printf("stop In\n")
				return
			}
		default:
			if m.onInPut != nil {
				b := m.onInPut()
				if b == nil {
					fmt.Printf("close In\n")
					return
				}
				for _, nal := range b {
					m.streamQueue <- nal.Buf
				}
			} else {
				return
			}
		}
	}
}

func (m *MemberNode) runLoopDec() {
	if !m.IsNeedIn {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	for {
		select {
		case <-m.Ctx.Done():
			{
				//调用cancel了
				fmt.Printf("stop dec\n")
				return
			}
		case rawStream := <-m.streamQueue: //裸流
			{
				//收到数据流
				if rawStream == nil {
					//说明协议层关闭
					return
				}
				av := m.InBuffer.Current()
				dataBuf, ok := av.Value.(*libvideo.LvcFramePicDataSt)
				if !ok {
					continue
				}

				ok = libvideo.LVC_VideoDecoderProcFrameNoLine(m.lvcDec, dataBuf, rawStream)
				//fmt.Printf("decode succ:%v\n", ok)
				if ok {
					if m.subInRing == nil {
						rr := m.InBuffer.Ring
						m.subInRing = m.InBuffer.SubRing(rr)
					}
				}
				m.InBuffer.Step()
			}
		}
	}
}

func (m *MemberNode) runLoopOut() {
	if !m.IsNeedOut {
		return
	}
	m.wg.Add(1)
	defer m.wg.Done()
	for {
		select {
		case <-m.Ctx.Done():
			{
				//调用cancel了
				fmt.Printf("stop enc\n")
				return
			}
		default:
			if m.OutBuffer == nil {
				m.OutBuffer = m.Conference.OutBuffer.SubRing(m.Conference.OutBuffer.Ring)
			}
			av, _ := m.OutBuffer.TryRead()
			if av == nil {
				time.Sleep(time.Millisecond * 5)
				continue
			}
			dataBuf, ok := av.Value.([]byte)
			if !ok {
				continue
			}
			//fmt.Printf("get enc dd %d\n", av.Index)
			m.OutBuffer.Ring = m.OutBuffer.Ring.Next()
			if m.onOutPut != nil {
				m.onOutPut(dataBuf, uint32(m.ts))
			}
		}
	}
}

func (m *MemberNode) StopMember() bool {
	m.memLock.Lock()
	defer m.memLock.Unlock()
	if m.cancel == nil {
		return false
	}

	if m.onClose != nil {
		m.onClose()
	}

	m.cancel()

	m.wg.Wait()

	if m.onCLoseDelay != nil {
		m.onCLoseDelay()
	}

	m.cancel = nil
	//通知conference
	if m.lvcDec != nil {
		libvideo.LVC_FreeDecoder(m.lvcDec)
	}
	close(m.streamQueue)
	m.streamQueue = nil
	if m.Conference == nil {
		return true
	}
	m.Conference = nil
	m.InBuffer.StopRing()
	m.InBuffer = nil
	return true
}
