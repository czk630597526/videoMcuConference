package demoConference

import (
	"McuConference/conference"
	"fmt"
	"os"
	"time"
)

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
