package client

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"tcp-serv-test/internal/message"
)

const (
	headerTypeNewClient = iota
	headerTypeClientList
	headerTypeDisconnectClient
	headerTypeClientMessage
)

type messageContent struct {
	headerType int
	content    string
}

// Client tcp chat client
type Client struct {
	address string
	clients map[string]bool
	conn    net.Conn
	stops   bool
}

// New creates new Client
func New(address string) *Client {
	return &Client{
		address: address,
		clients: map[string]bool{},
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
			m, err := message.Encode(fmt.Sprintf("%s%s", "[client-message]", strings.Trim(input, "\n ")))
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

		messageVal, err := c.getMessageVal(content)
		if err != nil {
			log.Println("unexpected message format")
		}

		switch messageVal.headerType {
		case headerTypeNewClient:
			c.clients[content] = true
			content = "new client: " + messageVal.content
		case headerTypeClientList:
			c.clients[content] = true
			content = "existed client: " + messageVal.content
		case headerTypeDisconnectClient:
			delete(c.clients, content)
			content = "client disconnected: " + messageVal.content
		case headerTypeClientMessage:
			content = messageVal.content
		}

		fmt.Println(content)
	}
}

func (c Client) getMessageVal(content string) (messageContent, error) {
	if strings.HasPrefix(content, "[new-client]") {
		return messageContent{
			headerTypeNewClient,
			strings.TrimPrefix(content, "[new-client]"),
		}, nil
	}
	if strings.HasPrefix(content, "[clients-list]") {
		return messageContent{
			headerTypeClientList,
			strings.TrimPrefix(content, "[clients-list]"),
		}, nil
	}
	if strings.HasPrefix(content, "[client-disconnect]") {
		return messageContent{
			headerTypeDisconnectClient,
			strings.TrimPrefix(content, "[client-disconnect]"),
		}, nil
	}
	if strings.HasPrefix(content, "[client-message]") {
		return messageContent{
			headerTypeClientMessage,
			strings.TrimPrefix(content, "[client-message]"),
		}, nil
	}
	return messageContent{}, errors.New("wrong message format")
}

// Stop stops chat client
func (c *Client) Stop() {
	c.stops = true
	_ = c.conn.Close()
}
