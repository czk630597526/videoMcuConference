package conference

import "McuConference/libvideo_codec"

type CreateConfInfo struct {
	Name      string `json:"name"`
	Pwd       string `json:"pwd"`
	HdrString string `json:"hdrString"`
	Rate      int    `json:"rate"`
	Bitrate   int    `json:"bitrate"`
}

type UserNodeCallBak struct {
	FuncOnCloseDelay func()
	FuncOnClose      func()
	FuncOnInput      func() []*libvideo_codec.H264Nal
	FuncOnOutput     func([]byte, uint32)
}

type DeleteConfInfo struct {
	Name string `json:"name"`
}

type UpdateInfo struct {
	RoomName string `json:"roomName"`
	NodeName string `json:"nodeName"`
	Index    int    `json:"index"`
}

type JoinInfo struct {
	RoomName    string `json:"roomName"`
	NodeName    string `json:"nodeName"`
	IsIn        bool   `json:"isIn"`
	IsOut       bool   `json:"isOut"`
	RemoteVAddr string `json:"remoteVAddr"`
}

type LeaveInfo struct {
	RoomName string `json:"roomName"`
	NodeName string `json:"nodeName"`
}

type MediaInfo struct {
	LocalVAddr string `json:"localVPort"`
}

type RspInfo struct {
	Method   string      `json:"method"`
	ErrCode  int         `json:"errcode"`
	RoomName string      `json:"roomName"`
	NodeName string      `json:"nodeName"`
	ErrMsg   string      `json:"errMsg"`
	Param    interface{} `json:"param,omitempty"`
}

type ReqInfo struct {
	Method string      `json:"method"`
	Param  interface{} `json:"param"`
}
