# heartbeat
Implement a simple hearbeat detect with HTTP protocol.

For old version which use UDP protocol. see [tag 1.0](#TODO)

## Install
```bash
go get -v github.com/codeskyblue/heartbeat
```

## Usage
```
```

### 服务端
	func main(){
		watcher := heartbeat.NewWatcher(":7788")
		notify, err := watcher.Watch()
		if err != nil {
			fmt.Println(err)
			return
		}
		
		for {
			// 当机器状态变化的时候，管道(notify)就会接收到消息
			host := <-notify
			fmt.Println(host.IP, host.Alive)
			fmt.Println(string(watcher.JSONMessage()))
		}
	}
### 客户端
	func main(){
		//heartbeat.DEFAULTGAP = time.Second * 2 // 调整发送间隔
		heartbeat.GoBeat("localhost:7788") // 向localhost发送消息
	}
	

# LICENSE
[GNU 2.0](LICENSE)