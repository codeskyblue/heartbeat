package heartbeat

import (
	"github.com/shxsun/beelog"
	"net"
	"time"
)

var DEFAULTMSG = "imok"
var DEFAULTGAP = time.Second

func GoBeat(addr string) {
	go func() {
		for {
			Beat(addr, DEFAULTMSG)
			time.Sleep(DEFAULTGAP)
		}
	}()
}

// tell the master I'am alive
func Beat(addr string, msg string) {
	conn, err := net.DialTimeout("udp", addr, time.Second*2)
	if err != nil {
		beelog.Error(err)
		return
	}
	defer conn.Close()
	beelog.Debug("beat send msg:", msg)
	_, err = conn.Write([]byte(msg))
	return
}
