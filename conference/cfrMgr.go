package conference

import (
	"fmt"
	"sync"
)

type ConferenceMgr struct {
	mapLock sync.RWMutex
	NameMap map[string]*ConferenceNode
}

var instance *ConferenceMgr

func CreateCfrMgr() *ConferenceMgr {
	if instance != nil {
		return instance
	}
	c := &ConferenceMgr{
		NameMap: make(map[string]*ConferenceNode),
	}
	instance = c
	return c
}

func DeleteCfrMgr(cfr *ConferenceMgr) {
	instance = nil
}

func (mgr *ConferenceMgr) AddConferenceNode(info *CreateConfInfo) error {
	mgr.mapLock.Lock()
	defer mgr.mapLock.Unlock()
	_, ok := mgr.NameMap[info.Name]
	if ok {
		return fmt.Errorf("dup conference name")
	}
	cfr := CreateConference(info.Name, info.Pwd, info.HdrString, info.Rate, info.Bitrate)
	if cfr == nil {
		return fmt.Errorf("create conference error")
	}
	err := cfr.StartConference()
	if err != nil {
		return fmt.Errorf("start conference error %s", err.Error())
	}
	mgr.NameMap[info.Name] = cfr
	return nil
}

func (mgr *ConferenceMgr) DelConferenceNode(info *DeleteConfInfo) error {
	mgr.mapLock.Lock()
	defer mgr.mapLock.Unlock()
	c := mgr.NameMap[info.Name]
	if c == nil {
		return fmt.Errorf("can not find conference:%s\n", info.Name)
	}
	c.StopConference()
	delete(mgr.NameMap, info.Name)

	return nil
}

func (mgr *ConferenceMgr) UpdateMemberNode(info *UpdateInfo) error {
	mgr.mapLock.RLock()
	defer mgr.mapLock.RUnlock()
	c := mgr.NameMap[info.RoomName]
	if c == nil {
		return fmt.Errorf("can not find conference:%s\n", info.RoomName)
	}

	e := c.UpdateNodeIndex(info.NodeName, info.Index)
	if e != nil {
		return e
	}
	return nil
}

func (mgr *ConferenceMgr) JoinMemberNode(info *JoinInfo, cb UserNodeCallBak) error {
	mgr.mapLock.RLock()
	defer mgr.mapLock.RUnlock()
	c := mgr.NameMap[info.RoomName]
	if c == nil {
		return fmt.Errorf("can not find conference:%s\n", info.RoomName)
	}
	n := CreateMember(info.NodeName)
	if n == nil {
		return fmt.Errorf("create node error:%s\n", info.NodeName)
	}
	n.SetInBuffer(info.IsIn)
	n.SetOutBuffer(info.IsOut)
	n.onCLoseDelay = cb.FuncOnCloseDelay
	n.onClose = cb.FuncOnClose
	n.onInPut = cb.FuncOnInput
	n.onOutPut = cb.FuncOnOutput

	e := c.AddMemberNode(n)
	if e != nil {
		return e
	}
	n.StartMember()

	return nil
}

func (mgr *ConferenceMgr) LeaveMemberNode(info *LeaveInfo) error {
	mgr.mapLock.RLock()
	defer mgr.mapLock.RUnlock()
	c := mgr.NameMap[info.RoomName]
	if c == nil {
		return fmt.Errorf("can not find conference:%s\n", info.RoomName)
	}

	e := c.DelMemberNode(info.NodeName)
	if e != nil {
		return e
	}
	return nil
}
