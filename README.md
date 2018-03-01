# heartbeat
Implement a simple hearbeat detect with HTTP protocol with secret.

For old version which use UDP protocol. see [tag 1.0](#TODO)

## Install
```bash
go get -v github.com/codeskyblue/heartbeat
```

## Usage
Server Example:

```go
hbs := NewServer("my-secret", 15 * time.Second) // secret: my-secret, timeout: 15s
hbs.OnConnect = func(identifier string) {
	fmt.Println(identifier, "is online")
}
hbs.OnDisconnect = func(identifier string) {
	fmt.Println(identifier, "is offline")
}
http.Handle("/heartbeat", hbs)
```

Client Example:

```go
cancel := &Client{
	ServerAddr: "http://hearbeat.example.com/heartbeat",
	Secret: "my-secret", // must be save with server secret
	Identifier: "client-unique-name",
}.Beat(5 * time.Second)

defer cancel() // cancel heartbeat
```

## Protocol
1. client get timestamp from server
2. client send identifier, timestamp and hmac hash to server every interval
3. server send back the new timestamp to client on each request

# LICENSE
[GNU 2.0](LICENSE)