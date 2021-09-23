package server

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	msg "tcp-serv-test/internal/message"
)

func TestServer_Serve(t *testing.T) {
	address := ":8081"
	s := New(address)
	go s.Serve()
	defer s.Stop(context.Background())

	conn1, err := buildClient(address)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	conn2, err := buildClient(address)
	if err != nil {
		t.Fatal(err)
	}

	// assert connection messages
	_ = conn1.SetReadDeadline(time.Now().Add(10 * time.Second))
	m, err := msg.Read(conn1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(m[2:]), "[new-client]") {
		t.Fatal("conn1 didn't get new-client header from conn2")
	}
	conn2ID := strings.TrimPrefix(string(m[2:]), "[new-client]")

	_ = conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	m, err = msg.Read(conn2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(m[2:]), "[clients-list]") {
		t.Fatalf("conn2 didn't get clients-list header from conn1, %s", m)
	}

	time.Sleep(1 * time.Second)
	conn3, err := buildClient(address)
	if err != nil {
		t.Fatal(err)
	}

	// reading conn3 connection messages
	_ = conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, err = msg.Read(conn2)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn3.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, err = msg.Read(conn3)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn3.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, err = msg.Read(conn3)
	if err != nil {
		t.Fatal(err)
	}

	// broadcast
	expectedMsg := []byte{0, 1, 'A'}
	_, err = conn1.Write(expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	_ = conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	m, err = msg.Read(conn2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, expectedMsg) {
		t.Fatalf("message is not equal with expected\nActual:   %b\nExpected: %b", m, expectedMsg)
	}

	_ = conn3.SetReadDeadline(time.Now().Add(10 * time.Second))
	m, err = msg.Read(conn3)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, expectedMsg) {
		t.Fatalf("message is not equal with expected\nActual:   %b\nExpected: %b", m, expectedMsg)
	}

	// direct
	expectedMsg, _ = msg.Encode("@" + conn2ID + "test msg")
	_, err = conn1.Write(expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	_ = conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	m, err = msg.Read(conn2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, expectedMsg) {
		t.Fatalf("message is not equal with expected\nActual:   %b\nExpected: %b", m, expectedMsg)
	}

	_ = conn3.SetReadDeadline(time.Now().Add(2 * time.Second))
	m, err = msg.Read(conn3)
	if err == nil {
		t.Fatalf("conn3 received message directed to conn2, message: %b", m)
	}
	if !strings.Contains(err.Error(), "i/o timeout") {
		t.Fatalf("unexpected error %e", err)
	}
}

func buildClient(address string) (net.Conn, error) {
	var err error
	for i := 0; i < 10; i++ {
		var conn net.Conn
		conn, err = net.Dial("tcp", address)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		tcpConn := conn.(*net.TCPConn)
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		return tcpConn, nil
	}
	return nil, err
}
