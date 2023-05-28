package config

import (
	"McuConference/liblog"
	"McuConference/udpmedia"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//Config global
var Config = loadConfig()

//ConfigST struct
type ConfigST struct {
	Server ServerST `json:"server"`
}

//ServerST struct
type ServerST struct {
	ServerIp  string `json:"server_ip"`
	StartPort int    `json:"udp_start_port"`
	EndPort   int    `json:"udp_end_port"`
	LogLevel  int    `json:"log_level"`
	StunUrl   string `json:"stun_url"`
}

func ConfigInit() {
	loadConfig()
}

func loadConfig() *ConfigST {
	var tmp ConfigST
	data, err := ioutil.ReadFile("/etc/manager/conference.json")
	if err != nil {
		data, err = ioutil.ReadFile("conference.json")
		if err != nil {
			//liblog.Log.Elog("%v", err)
			fmt.Println(err)
			os.Exit(0)
			//return nil
		}
	}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		//liblog.Log.Elog("%v", err)
		fmt.Println("parse config json fail: ", err)
		os.Exit(0)
	}

	if tmp.Server.StartPort >= tmp.Server.EndPort {
		fmt.Println("udp_start_port >= udp_end_port, check webrtcMedia.json")
		os.Exit(0)
	}

	liblog.Init_log(tmp.Server.LogLevel, true)

	udpmedia.MediaInit(tmp.Server.StartPort, tmp.Server.EndPort, tmp.Server.ServerIp)

	return &tmp
}
