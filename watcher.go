// Heartbeat
/*
机器的存活是根据HP来的。
HP越高，说明机器活跃度越高

*/
package heartbeat

import (
	"encoding/json"
	"github.com/shxsun/beelog"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type Host struct {
	IP    string
	Time  time.Time
	HP    int
	Alive bool
}

var (
	LevelInitHP    int = 3  // 接收到心跳时，HP最低为该值
	LevelAliveHP   int = 7  // 复活时的HP
	LevelFullHP    int = 10 // 满血时的HP
	DEFAULTRECYCLE     = time.Second * 2
)

func init() {
	beelog.SetLevel(beelog.LevelInfo)
}

// use udp protocal
type Watcher struct {
	sync.Mutex
	Addr            string
	RecycleDuration time.Duration
	Hosts           map[string]*Host
}

func NewWatcher(addr string) *Watcher {
	return &Watcher{
		Addr:            addr,
		RecycleDuration: DEFAULTRECYCLE,
		Hosts:           make(map[string]*Host, 100),
	}
}

// cut down HP
func (this *Watcher) hurt(ip string) {
	beelog.Trace("hurt", ip)
	h, ok := this.Hosts[ip]
	if !ok {
		return
	}
	if h.HP -= 1; h.HP < 0 {
		h.HP = 0
	}
	this.updateState(ip)
}

// recover HP
func (this *Watcher) fix(ip string) {
	this.Lock()
	defer this.Unlock()
	beelog.Trace("fix", ip)
	h, ok := this.Hosts[ip]
	if !ok {
		this.Hosts[ip] = &Host{IP: ip, Time: time.Now(), HP: LevelFullHP, Alive: true}
		beelog.Trace("Len hosts:", len(this.Hosts))
		return
	}
	h.HP += 1
	if h.HP > LevelFullHP {
		h.HP = LevelFullHP
	}
	if h.HP < LevelInitHP {
		h.HP = LevelInitHP
	}
	h.Time = time.Now()
	this.updateState(ip)
}

// judge if host is Alive from HP
func (this *Watcher) updateState(ip string) {
	host, ok := this.Hosts[ip]
	if !ok {
		return
	}
	if host.HP >= LevelAliveHP {
		host.Alive = true
	}
	if host.HP == 0 {
		host.Alive = false
	}
}

// return {"Alives": [...], "Deads": [...]}
func (this *Watcher) JSONMessage() []byte {
	this.Lock()
	defer this.Unlock()
	alives := make([]string, 0, len(this.Hosts))
	deads := make([]string, 0, len(this.Hosts))
	for ip, host := range this.Hosts {
		if host.Alive {
			alives = append(alives, lookupName(ip))
		} else {
			deads = append(deads, lookupName(ip))
		}
	}
	sort.Strings(alives)
	sort.Strings(deads)
	data, err := json.Marshal(struct {
		Alives []string
		Deads  []string
	}{
		alives,
		deads,
	})
	if err != nil {
		panic(err)
	}
	return data
}

// When host is dead or come to alive, chan calls.
func (this *Watcher) Watch() (chan Host, error) {
	ch, err := this.listen()
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			ip := <-ch
			beelog.Trace("read channel ip:", ip)
			this.fix(ip)
		}
	}()

	notify := make(chan Host, 50)
	go this.drain(notify) // clean program
	return notify, nil
}

// auto decrease host HP
func (this *Watcher) drain(notify chan Host) {
	beelog.Debug("drain stated")
	for {
		this.Lock()
		for _, host := range this.Hosts {
			var state = host.Alive
			this.hurt(host.IP)
			beelog.Debug("drain", host.IP, host.HP, host.Alive)
			if host.Alive != state {
				notify <- *host
			}
		}
		this.Unlock()
		time.Sleep(this.RecycleDuration)
	}
}

// Listen UDP packet and write to channel
func (this *Watcher) listen() (chan string, error) {
	lis, err := net.ListenPacket("udp", this.Addr)
	if err != nil {
		return nil, err
	}
	beelog.Info("start listen packet from", this.Addr)

	ch := make(chan string)
	go func() {
		defer lis.Close()
		buf := make([]byte, 1000)
		for {
			n, addr, err := lis.ReadFrom(buf)
			msg := string(buf[:n])
			beelog.Debug("Receive from", addr, msg)
			if err != nil {
				beelog.Warn(err)
				continue
			}
			ip := strings.Split(addr.String(), ":")[0]
			ch <- ip
		}
	}()
	return ch, nil
}
