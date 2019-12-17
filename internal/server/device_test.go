package server

import (
	"io"
	"net"
	"testing"
	"time"
)

func Test_Device_ReadDeadline_Positive(t *testing.T) {

	// test server
	testAddr := ":7373"
	ln, err := net.Listen("tcp", testAddr)
	if err != nil {
		t.Fatalf("new listener err: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("accept new connect, err: %v", err)
		}

		ld := time.Millisecond * 50
		md := time.Millisecond * 50
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn)
		err = d.run()
		if err == io.EOF {
			t.Logf("test server get EOF")
		} else if err != nil {
			t.Fatalf("device run err: %v", err)
		}
	}()

	// test device deadlines
	// send login, message
	clnConn, err := net.Dial("tcp", testAddr)
	if err != nil {
		t.Fatalf("client conn err: %v", err)
	}

	// send login
	imei := make([]byte, imeiLength)
	_, err = clnConn.Write(imei)
	if err != nil {
		t.Fatalf("client conn write imei err: %v", err)
	}

	// send message
	msg := make([]byte, msgLength)
	_, err = clnConn.Write(msg)
	if err != nil {
		t.Fatalf("client conn write message err: %v", err)
	}

	// close client connection
	err = clnConn.Close()
	if err != nil {
		t.Fatalf("client conn close err: %v", err)
	}
	time.Sleep(time.Millisecond * 100)
}

func Test_Device_LoginDeadline_Negative(t *testing.T) {

	// test server
	testAddr := ":7374"
	ln, err := net.Listen("tcp", testAddr)
	if err != nil {
		t.Fatalf("new listener err: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("accept new connect, err: %v", err)
		}

		ld := time.Millisecond * 50
		md := time.Millisecond * 50
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn)
		err = d.run()
		if e, ok := err.(net.Error); ok && e.Timeout() {
			t.Logf("test server get i/o timeout")
		} else if err != nil {
			t.Fatalf("device run err: %v", err)
		}
	}()

	// test device deadlines
	// send login, message
	clnConn, err := net.Dial("tcp", testAddr)
	if err != nil {
		t.Fatalf("client conn err: %v", err)
	}

	// send login with delay
	time.Sleep(time.Millisecond * 100)
	err = clnConn.Close()
	if err != nil {
		t.Fatalf("client conn close err: %v", err)
	}
}

func Test_Device_MessageDeadline_Negative(t *testing.T) {

	// test server
	testAddr := ":7375"
	ln, err := net.Listen("tcp", testAddr)
	if err != nil {
		t.Fatalf("new listener err: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("accept new connect, err: %v", err)
		}

		ld := time.Millisecond * 50
		md := time.Millisecond * 50
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn)
		err = d.run()
		if e, ok := err.(net.Error); ok && e.Timeout() {
			t.Logf("test server get i/o timeout")
		} else if err != nil {
			t.Fatalf("device run err: %v", err)
		}
	}()

	// test device deadlines
	// send login, message
	clnConn, err := net.Dial("tcp", testAddr)
	if err != nil {
		t.Fatalf("client conn err: %v", err)
	}

	// send login
	imei := make([]byte, imeiLength)
	_, err = clnConn.Write(imei)
	if err != nil {
		t.Fatalf("client conn write imei err: %v", err)
	}

	// send message with delay
	time.Sleep(time.Millisecond * 100)
	err = clnConn.Close()
	if err != nil {
		t.Fatalf("client conn close err: %v", err)
	}
}
