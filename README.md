# heartbeat
Implement a simple hearbeat detect with HTTP protocol with secret.

For old version which use UDP protocol. see [tag 1.0](#TODO)

## Install
```bash
go get -v github.com/codeskyblue/heartbeat
```

## Usage
Server and Client should have the same secret.

Server Example:

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codeskyblue/heartbeat"
)

func main() {
	hbs := heartbeat.NewServer("my-secret", 15*time.Second) // secret: my-secret, timeout: 15s
	hbs.OnConnect = func(identifier string) {
		fmt.Println(identifier, "is online")
	}
	hbs.OnDisconnect = func(identifier string) {
		fmt.Println(identifier, "is offline")
	}
	http.Handle("/heartbeat", hbs)
	http.ListenAndServe(":7000", nil)
}
```

Client Example:

```go
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
```

## Protocol
1. client get timestamp from server
2. client send identifier, timestamp and hmac hash to server every interval
3. server send back the new timestamp to client on each request

# LICENSE
[GNU 2.0](LICENSE)