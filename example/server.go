package main

import (
	"flag"
	"fmt"
	"github.com/shxsun/jetfire/heartbeat"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

var watcher = heartbeat.NewWatcher(":7788")

var html = `
<html>
<head>
</head>
<body>
    <h1>Heart Beat</h1>
    <h2>Alives</h2>
    <ul>
    {{range .Alives}} <li>{{.}}</li> {{end}}
    </ul>
    <hr/>
    <h2>Deads</h2>
    <ul>
    {{range .Deads}} <li>{{.}}</li> {{end}}
    </ul>
    <p></p>
    <p>open api <a href="/json">REST API</a></p>
</body>
</html>
`

// FIXME
func serveHttp(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := template.Must(template.New("heart").Parse(html)).Execute(w, watcher.JSONMessage())
		if err != nil {
			fmt.Println(err)
		}
	})

	http.HandleFunc("/json/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(watcher.JSONMessage())
	})
	http.ListenAndServe(addr, nil)
}

func main() {
	mail := flag.String("mail", "sunshengxiang01@baidu.com", "email alert address")
	flag.Parse()

	notify, err := watcher.Watch()
	if err != nil {
		fmt.Println(err)
		return
	}
	go serveHttp(":8120")
	for {
		h := <-notify
		fmt.Println(h.IP, h.Alive)
		fmt.Println(string(watcher.JSONMessage()))
		status := "dead"
		if h.Alive {
			status = "come to alive"
		}

		var hname string
		hs, err := net.LookupAddr(h.IP)
		if err != nil {
			fmt.Println(err)
			hname = h.IP
		} else {
			for _, s := range hs {
				if strings.Contains(s, "baidu.com") {
					hname = s
					break
				}
			}
		}

		subject := fmt.Sprintf("[heartbeat] <%s> %s", hname, status)
		cmd := exec.Command("mutt", "-s", subject, *mail)
		cmd.Stdin = strings.NewReader(
			fmt.Sprintf("machine [%s] %s. \n\nps: this is an auto mail. don't need reply", hname, status))
		er := cmd.Start()
		if er != nil {
			fmt.Println(er)
		}
	}
	fmt.Println("END")
}
