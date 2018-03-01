package heartbeat

import (
	"log"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeartbeat(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	hbs := NewServer("kitty", 4*time.Second)
	hbs.OnConnect = func(identifier string) {
		log.Println("connect", identifier)
	}
	hbs.OnDisconnect = func(identifier string) {
		log.Println("disconnect", identifier)
	}
	ts := httptest.NewServer(hbs)
	defer ts.Close()

	log.Println(ts.URL)
	client := &Client{
		Secret:     "kitty2",
		Identifier: "whoami",
		ServerAddr: ts.URL,
	}
	cancel := client.Beat(2 * time.Second)
	time.Sleep(3e9)
	cancel()
	time.Sleep(5e9)
	log.Println("FINISHED")
}
