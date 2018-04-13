/*
Protocols between client and server

Client (HTTP Request) ->

	Query: identifier (uniq string)
	Query: timestamp (seconds since January 1, 1970 UTC.)
	Query: hashmac

Server response ->
	Body: {timestamp} {hashmac}
*/
package heartbeat

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Server struct {
	OnConnect    func(identifier string, req *http.Request)
	OnReconnect  func(identifier string, req *http.Request)
	OnDisconnect func(identifier string)
	hbTimeout    time.Duration
	secret       string // HMAC
	sessions     map[string]*Session
	mu           sync.Mutex
}

// NewServer accept secret, Client must have the same secret, so they can work together.
func NewServer(secret string, timeout time.Duration) *Server {
	return &Server{
		hbTimeout: timeout,
		secret:    secret,
		sessions:  make(map[string]*Session),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timestamp := r.FormValue("timestamp")
	identifier := r.FormValue("identifier")
	messageMAC := r.FormValue("messageMAC")

	if identifier == "" {
		http.Error(w, "identifier should not be empty", http.StatusBadRequest)
		return
	}
	// check hash MAC
	if messageMAC != hashIdentifier(timestamp, identifier, s.secret) {
		http.Error(w, "messageMAC wrong", http.StatusBadRequest)
		return
	}
	// check timestamp
	if timestamp != "" {
		var t int64
		fmt.Sscanf(timestamp, "%d", &t)
		if time.Now().Unix()-t < 0 || time.Now().Unix()-t > int64(s.hbTimeout.Seconds()) {
			http.Error(w, "Invalid timestamp, advanced or outdated", http.StatusBadRequest)
			return
		}
		go s.updateOrSaveSession(identifier, r)
	}

	// send server timestamp to client
	t := time.Now().Unix()
	fmt.Fprintf(w, "%d %s", t, hashTimestamp(fmt.Sprintf("%d", t), s.secret))
}

func (s *Server) updateOrSaveSession(identifier string, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	remoteHost, _, _ := net.SplitHostPort(req.RemoteAddr)
	if sess, ok := s.sessions[identifier]; ok {
		// Call OnReconnect again when client IP changes
		if sess.remoteHost != remoteHost {
			sess.remoteHost = remoteHost
			if s.OnReconnect != nil {
				s.OnReconnect(identifier, req)
			}
		}
		select {
		case sess.recvC <- "beat":
			// log.Println(sess.identifier, "beat")
		default:
		}
	} else {
		if s.OnConnect != nil {
			s.OnConnect(identifier, req)
		}
		sess := &Session{
			identifier: identifier,
			remoteHost: remoteHost,
			timer:      time.NewTimer(s.hbTimeout),
			timeout:    s.hbTimeout,
			recvC:      make(chan string, 0),
		}
		s.sessions[identifier] = sess
		go func() {
			sess.drain()
			// delete session when timeout
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.OnDisconnect != nil {
				s.OnDisconnect(identifier)
			}
			delete(s.sessions, identifier)
		}()
	}
}

type Session struct {
	identifier string
	remoteHost string
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

type Client struct {
	Secret     string
	Identifier string
	ServerAddr string
	OnConnect  func()
	OnError    func(error)
}

// Beat send identifier and hmac hash to server every interval
func (c *Client) Beat(interval time.Duration) (cancel context.CancelFunc) {
	if !regexp.MustCompile(`^https?://`).MatchString(c.ServerAddr) {
		c.ServerAddr = "http://" + c.ServerAddr
	}
	ctx, cancel := context.WithCancel(context.TODO())
	go func() {
		for {
			timeKey, err := c.httpBeat("", c.ServerAddr)
			if err != nil {
				sleepDuration := interval + time.Duration(rand.Intn(5))*time.Second
				// secret might wrong
				if strings.Contains(err.Error(), "messageMAC wrong") {
					sleepDuration += 1 * time.Minute
				}
				if c.OnError != nil {
					c.OnError(err)
				}
				log.Printf("heatbeat err: %v, retry after %v", err, sleepDuration)
				time.Sleep(sleepDuration)
				continue
			}
			if c.OnConnect != nil {
				c.OnConnect()
			}
			err = c.beatLoop(ctx, interval, timeKey)
			if err == nil {
				break
			}
		}
	}()
	return cancel
}

// send hearbeat continously
func (c *Client) beatLoop(ctx context.Context, interval time.Duration, timeKey string) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		newTimeKey, er := c.httpBeat(timeKey, c.ServerAddr)
		if er != nil {
			return errors.Wrap(er, "beatLoop")
		}
		timeKey = newTimeKey
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (c *Client) httpBeat(serverTimeKey string, serverAddr string) (timeKey string, err error) {
	httpclient := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := httpclient.PostForm(serverAddr, url.Values{ // TODO: http timeout
		"timestamp":  {serverTimeKey},
		"identifier": {c.Identifier},
		"messageMAC": {hashIdentifier(serverTimeKey, c.Identifier, c.Secret)}})
	if err != nil {
		err = errors.Wrap(err, "post form")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "ioutil readall")
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New(strings.TrimSpace(string(body)))
		return
	}

	// Receive server timestamp and check server hmac HASH
	var hashMAC string
	n, err := fmt.Sscanf(string(body), "%s %s", &timeKey, &hashMAC)
	if err != nil {
		log.Println(n, err)
		return
	}
	if hashTimestamp(timeKey, c.Secret) != hashMAC {
		err = errors.New("wrong timestamp hmac")
		return
	}
	return
}

func hashTimestamp(t, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:timestamp", t)))
	return hex.EncodeToString(mac.Sum(nil))
}

func hashIdentifier(timestamp, identifier, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%s", timestamp, identifier)))
	return hex.EncodeToString(mac.Sum(nil))
}
