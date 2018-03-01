package main

import (
	"time"

	"github.com/codeskyblue/heartbeat"
)

func main() {
	client := &heartbeat.Client{
		ServerAddr: "http://localhost:7000/heartbeat", // replace to your server addr
		Secret:     "my-secret",                       // must be save with server secret
		Identifier: "client-unique-name",
	}
	cancel := client.Beat(5 * time.Second)
	defer cancel() // cancel heartbeat
	// Do something else
	time.Sleep(10 * time.Second)
}
