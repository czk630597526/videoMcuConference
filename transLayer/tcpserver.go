package transLayer

import (
	"McuConference/conference"
	"fmt"
	"net"
)

func StartTcpServer(addr string) error {
	mgr := conference.CreateCfrMgr()
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen error:%s", err.Error())
	}
	for {
		conn, err := listen.Accept() // 监听客户端的连接请求
		if err != nil {
			fmt.Println("Accept() failed, err: ", err)
			continue
		}
		node := CreateTcpNode(conn)
		node.cfrMgr = mgr
		err = node.Start()
		if err != nil {
			conn.Close()
			continue
		}
	}
}
