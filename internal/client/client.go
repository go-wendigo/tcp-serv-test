package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"tcp-serv-test/internal/message"
	"time"
)

// Client tcp chat client
type Client struct {
	address string
	clients []string
	conn    net.Conn
	stops   bool
}

// New creates new Client
func New(address string) *Client {
	return &Client{
		address: address,
	}
}

// Start starts chat client
func (c *Client) Start() {
	addr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		log.Fatalf("wrong server address %s", c.address)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c.conn = conn

	err = conn.SetKeepAlive(true)
	if err != nil {
		log.Println(err)
		return
	}
	err = conn.SetKeepAlivePeriod(30 * time.Second)
	if err != nil {
		log.Println(err)
		return
	}

	notify := make(chan error)

	go c.listenMessages(notify)
	go c.listenInput(notify)

	err = <-notify
	if err != nil {
		log.Fatalf("connection dropped message: %s", err.Error())
	}
}

func (c *Client) listenInput(notify chan error) {
	func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				notify <- err
				return
			}
			m, err := message.Encode(strings.Trim(input, "\n "))
			if err != nil {
				notify <- err
				return
			}
			_, err = c.conn.Write(m)
			if err != nil {
				if c.stops {
					close(notify)
					return
				}
				notify <- err
				return
			}
		}
	}()
}

func (c *Client) listenMessages(notify chan error) {
	for {
		msg, err := message.Read(c.conn)
		if err != nil {
			if c.stops {
				close(notify)
				return
			}
			notify <- err
			return
		}
		content, err := message.Decode(msg)
		if err != nil {
			notify <- err
			return
		}
		if strings.HasPrefix(content, "[new-client]") || strings.HasPrefix(content, "[clients-list]") {
			headerParts := strings.Split(content, "]")
			if len(headerParts) > 1 {
				c.clients = append(c.clients, headerParts[1])
			}
		}
		fmt.Println(content)
	}
}

// Stop stops chat client
func (c *Client) Stop() {
	c.stops = true
	_ = c.conn.Close()
}
