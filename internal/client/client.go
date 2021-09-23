package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"tcp-serv-test/internal/message"
	"time"
)

// Client tcp chat client
type Client struct {
	address string
	clients []string
	lock    *sync.Mutex
	conn    net.Conn
}

// New creates new Client
func New(address string) *Client {
	return &Client{
		address: address,
		lock:    new(sync.Mutex),
	}
}

// Start starts chat client
func (c *Client) Start() {
	conn, _ := net.Dial("tcp", c.address)
	c.conn = conn

	err := conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		log.Println(err)
		return
	}
	err = conn.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		log.Println(err)
		return
	}

	notify := make(chan error)

	go func() {
		for {
			msg, err := message.Read(conn)
			if err != nil {
				notify <- err
			}
			content, err := message.Decode(msg)
			if err != nil {
				notify <- err
			}
			if strings.HasPrefix(content, "[new-client]") || strings.HasPrefix(content, "[clients-list]") {
				split := strings.Split(content, "]")
				c.lock.Lock()
				c.clients = append(c.clients, split[1])
				c.lock.Unlock()
			}
			fmt.Println(content)
		}
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				notify <- err
			}
			m, err := message.Encode(strings.Trim(input, "\n "))
			if err != nil {
				notify <- err
			}
			_, err = conn.Write(m)
			if err != nil {
				notify <- err
			}
		}
	}()

	err = <-notify
	log.Println("connection dropped message", err)
	return
}

// Stop stops chat client
func (c *Client) Stop() {
	_ = c.conn.Close()
}
