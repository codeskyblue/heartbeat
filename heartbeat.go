/*
Server Example:

	hbs := NewServer(15 * time.Second)
	hbs.OnConnect = func(identifier string, message string) {
		fmt.Println(identifier, "is online")
	}
	hbs.OnDisconnect = func(identifier string) {
		fmt.Println(identifier, "is offline")
	}
	http.Handle("/heartbeat", hbs)

Client Example:

	Beat("http://hearbeat.example.com/heartbeat", 5 *time.Second)
*/
package heartbeat

import (
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

const DefaultTimeout = 15 * time.Second

type Server struct {
	OnConnect    func(identifier string, message string)
	OnDisconnect func(identifier string)
	hbTimeout    time.Duration
	secret       string
	sessions     map[string]*Session
	mu           sync.Mutex
}

func NewServer(secret string, timeout time.Duration) *Server {
	return &Server{
		hbTimeout: DefaultTimeout,
		secret:    secret,
		sessions:  make(map[string]*Session),
	}
}

// hmac
func (s *Server) checkMAC(message, messageMAC []byte, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return hmac.Equal(messageMAC, mac.Sum(nil))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := r.Header.Get("X-Hash-Message")
	messageMAC := r.Header.Get("X-Hash-Result")
	if !s.checkMAC([]byte(message), []byte(messageMAC), s.secret) {
		// TODO
		return
	}

	identifier, _, _ := net.SplitHostPort(r.RemoteAddr)
	body, _ := ioutil.ReadAll(r.Body)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if sess, ok := s.sessions[identifier]; ok {
			select {
			case sess.recvC <- "beat":
			default:
			}
		} else {
			if s.OnConnect != nil {
				s.OnConnect(identifier, string(body))
			}
			sess := &Session{
				identifier: identifier,
				timer:      time.NewTimer(s.hbTimeout),
				timeout:    s.hbTimeout,
				recvC:      make(chan string, 0),
			}
			s.sessions[identifier] = sess
			go func() {
				sess.drain()
				if s.OnDisconnect != nil {
					s.OnDisconnect(identifier)
				}
			}()
		}
	}()
	io.WriteString(w, "Success")
}

type Session struct {
	identifier string
	timer      *time.Timer
	timeout    time.Duration
	recvC      chan string
}

func (sess *Session) drain() {
	for {
		select {
		case <-sess.recvC:
			sess.timer.Reset(sess.timeout)
		case <-sess.timer.C:
			return
		}
	}
}

// TODO
func Beat(serverAddr string, interval time.Duration) {

}
