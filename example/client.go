package main

import (
	"flag"
	"fmt"
	"github.com/shxsun/jetfire/heartbeat"
	"time"
)

var (
	addr = flag.String("addr", "db-testing-oped3001.db01:7788", "heartbeat center address")
)

func main() {
	flag.Parse()
	heartbeat.GoBeat(*addr)
	for {
		time.Sleep(1e9)
	}
	fmt.Println("END")
}
