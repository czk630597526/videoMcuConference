package transLayer

import (
	"McuConference/conference"
	"encoding/binary"
	"encoding/json"
)

func MakeRsp(info *conference.RspInfo) []byte {
	head := []byte{'w', 'e', 'r', 'x'}
	b, e := json.Marshal(info)
	if e != nil {
		return nil
	}
	l := uint16(len(b))
	binary.BigEndian.PutUint16(head[4:], l)
	head = append(head, b...)
	return head
}
