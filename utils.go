package heartbeat

import (
	"net"
	"strings"
)

func lookupName(ip string) string {
	var name string
	hs, err := net.LookupAddr(ip)
	if err != nil || len(hs) == 0 {
		name = ip
	} else {
		name = strings.Join(hs, ",")
	}
	return name
}
