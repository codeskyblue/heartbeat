package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codeskyblue/heartbeat"
)

func main() {
	hbs := heartbeat.NewServer("my-secret", 15*time.Second) // secret: my-secret, timeout: 15s
	hbs.OnConnect = func(identifier string, r *http.Request) {
		fmt.Println(identifier, "is online")
	}
	hbs.OnDisconnect = func(identifier string) {
		fmt.Println(identifier, "is offline")
	}
	http.Handle("/heartbeat", hbs)
	http.ListenAndServe(":7000", nil)
}
