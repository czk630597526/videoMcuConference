package conference

import (
	"McuConference/libvideo"
	"container/ring"
	"context"
	"runtime"
	"time"
)

type AVItem struct {
	Value   interface{}
	canRead bool
	Index   int
}

type AVRing struct {
	Ring    *ring.Ring
	Context context.Context
	Size    int
	poll    time.Duration
	Type    int //0表示picData	1表示byte
}

func (r *AVRing) Init(ctx context.Context, n int, t int) *AVRing {
	r.Ring = ring.New(n)
	r.Context = ctx
	r.Size = n
	r.Type = t
	i := 0
	for x := r.Ring; x.Value == nil; x = x.Next() {
		p := new(AVItem)
		p.Index = i
		i++
		if t == 0 {
			p.Value = libvideo.LvcFramePicData_Create(0, 0, 0)
		} else {
			p.Value = nil
		}
		x.Value = p
	}
	return r
}

func (r *AVRing) StopRing() {
	for x := r.Ring; x.Value == nil; x = x.Next() {
		p, ok := x.Value.(AVItem)
		if ok {
			picData, ok := p.Value.(*libvideo.LvcFramePicDataSt)
			if ok {
				libvideo.LvcFramePicData_Delete(picData)
			}
		}
	}
}

func (r AVRing) Clone() *AVRing {
	return &r
}

//这样其实相当于一个浅拷贝，只复制了一个指针的指向
func (r AVRing) SubRing(rr *ring.Ring) *AVRing {
	r.Ring = rr
	return &r
}

func (r *AVRing) Write(value interface{}) {
	last := r.Current()
	last.Value = value
	r.GetNext().canRead = false
	last.canRead = true
}

func (r *AVRing) Step() {
	last := r.Current()
	r.GetNext().canRead = false
	last.canRead = true
}

func (r *AVRing) wait() {
	if r.poll == 0 {
		runtime.Gosched()
	} else {
		time.Sleep(r.poll)
	}
}

func (r *AVRing) CurrentValue() interface{} {
	return r.Current().Value
}

func (r *AVRing) Current() *AVItem {
	return r.Ring.Value.(*AVItem)
}

func (r *AVRing) NextValue() interface{} {
	return r.Ring.Next().Value.(*AVItem).Value
}
func (r *AVRing) PreItem() *AVItem {
	return r.Ring.Prev().Value.(*AVItem)
}

func (r *AVRing) GetNext() *AVItem {
	r.Ring = r.Ring.Next()
	return r.Current()
}

func (r *AVRing) Read() (item *AVItem, value interface{}) {
	current := r.Current()
	for r.Context.Err() == nil && !current.canRead {
		r.wait()
	}
	return current, current.Value
}

func (r *AVRing) TryRead() (item *AVItem, value interface{}) {
	current := r.Current()
	if r.Context.Err() == nil && !current.canRead {
		return nil, nil
	}
	return current, current.Value
}
