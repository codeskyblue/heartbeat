# heartbeat
心跳检测(采用UDP协议)

## 背景
本来是用在jetfire中的模块，用来检测jetfire进程是否存活的。踩了几个坑，花了些时间，现在我把它开源出来，给大家用了。

遇到的问题主要是机器偶尔发个心跳，就认为机器复活了。这样是不合理的，心跳必须持续的发上几次，才能认为进程复活了。
死亡也必是连续几次心跳接收不到才行。

## 使用方法
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
	
