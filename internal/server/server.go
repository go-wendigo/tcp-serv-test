package server

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"sync"

	msg "tcp-serv-test/internal/message"

	uuid "github.com/satori/go.uuid"
)

// Server tcp chat server
type Server struct {
	listener net.Listener
	address  string
	connMap  sync.Map
	messages chan *message
	group    *sync.WaitGroup
}

// New creates new Server
func New(address string) *Server {
	return &Server{
		address:  address,
		connMap:  sync.Map{},
		messages: make(chan *message, 1000),
		group:    new(sync.WaitGroup),
	}
}

type message struct {
	author    string
	recipient string
	data      []byte
}

// Serve starts server
func (s *Server) Serve() {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		panic(err.Error())
	}
	s.listener = l
	go s.sendMessages()
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		connID := uuid.NewV4().String()
		s.connMap.Store(connID, conn)
		err = s.notifyNewClient(connID)
		if err != nil {
			log.Printf("can't init connection %q", conn.RemoteAddr().String())
			return
		}
		go s.handleConnection(connID, conn)
	}
}

// Stop stops server, closes connections
func (s *Server) Stop(ctx context.Context) {
	done := make(chan struct{})

	go func() {
		s.group.Wait()
		close(done)
	}()

	s.connMap.Range(func(connID, value interface{}) bool {
		conn, ok := value.(net.Conn)
		if ok {
			_ = conn.Close()
		}
		return true
	})

	select {
	case <-ctx.Done():
		if s.listener != nil {
			_ = s.listener.Close()
		}
	case <-done:
		if s.listener != nil {
			_ = s.listener.Close()
		}
	}
}

func (s *Server) handleConnection(connID string, conn net.Conn) {
	s.group.Add(1)
	log.Printf("serving %q - %q\n", conn.RemoteAddr().String(), connID)
	defer func() {
		log.Printf("closing connection %q\n", conn.RemoteAddr().String())
		_ = conn.Close()
		s.connMap.Delete(connID)
		s.group.Done()
	}()

	reader := bufio.NewReader(conn)
	for {
		data, err := msg.Read(reader)
		if err != nil {
			log.Printf("wrong message format from %q\n", conn.RemoteAddr().String())
			break
		}

		m := &message{
			author: connID,
			data:   data,
		}
		if recipient := s.getRecipient(data); recipient != "" {
			m.recipient = recipient
		}
		s.messages <- m
	}
}

func (s *Server) sendMessages() {
	for message := range s.messages {
		message := message
		s.connMap.Range(func(key, value interface{}) bool {
			connID, ok := key.(string)
			if !ok {
				return true
			}

			if connID == message.author ||
				message.recipient != "" && message.recipient != connID {
				return true
			}

			conn, ok := value.(net.Conn)
			if !ok {
				log.Printf("can't send message to %q, connection is failed", connID)
			}
			if _, err := conn.Write(message.data); err != nil {
				log.Printf("can't send message to %q", connID)
			}
			return true
		})
	}
}

func (s *Server) getRecipient(m []byte) string {
	if len(m) < 3 || m[2] != '@' {
		return ""
	}

	recipientID := uuid.FromStringOrNil(string(m[3:39]))
	if recipientID == uuid.Nil {
		return ""
	}
	return recipientID.String()
}

func (s *Server) notifyNewClient(connID string) error {
	newClientHeader, err := msg.Encode("[new-client]" + connID)
	if err != nil {
		return errors.New("can't notify about new client")
	}
	s.messages <- &message{
		author: connID,
		data:   newClientHeader,
	}

	s.connMap.Range(func(id, value interface{}) bool {
		if id == connID {
			return true
		}
		clientHeader, err := msg.Encode("[clients-list]" + id.(string))
		if err != nil {
			return true
		}
		s.messages <- &message{
			recipient: connID,
			data:      clientHeader,
		}
		return true
	})

	return nil
}
