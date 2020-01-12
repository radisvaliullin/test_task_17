package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	// 15-bytes decimal numbers
	testIMEI    = []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	testIMEIArr = [15]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	testOutLog  = log.New(os.Stdout, "", 0)
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
		wg := sync.WaitGroup{}
		wg.Add(1)
		stop := make(chan struct{}, 1)
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn, testOutLog, &wg, stop, newDevStorage())
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
	_, err = clnConn.Write(testIMEI)
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
		wg := sync.WaitGroup{}
		wg.Add(1)
		stop := make(chan struct{}, 1)
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn, testOutLog, &wg, stop, newDevStorage())
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
		wg := sync.WaitGroup{}
		wg.Add(1)
		stop := make(chan struct{}, 1)
		d := newDevice(devConfig{loginDeadline: ld, messageDeadline: md}, conn, testOutLog, &wg, stop, newDevStorage())
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
	_, err = clnConn.Write(testIMEI)
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

func Test_isValidIMEI(t *testing.T) {

	testCases := []struct {
		name string
		imei []byte
		err  error
	}{
		// Positive
		{
			name: "valid imei",
			imei: []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8},
			err:  nil,
		},
		// Negative
		{
			name: "invalid imei, wrong check",
			imei: []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 9},
			err:  errors.New("imei wrong check"),
		},
		{
			name: "invalid imei, wrong length",
			imei: []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1},
			err:  errors.New("imei wrong length"),
		},
		{
			name: "invalid imei, byte is not decimal",
			imei: []byte{49, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8},
			err:  errors.New("imei should be decimal number (0-9)"),
		},
	}

	for _, tc := range testCases {

		imei, err := validParseIMEI(tc.imei)
		if tc.err == nil && err != nil {
			t.Fatalf("%v: imei %v is invalid, err: %v", tc.name, tc.imei, err)
		} else if tc.err != nil && tc.err.Error() != err.Error() {
			t.Fatalf("%v: imei %v, negative test case return wrong error: tc.err - %v, err - %v", tc.name, tc.imei, tc.err, err)
		}
		t.Logf("%v: imei %v, imei string %v, test ok", tc.name, tc.imei, imei)
	}
}

func Test_bytesToFloat64(t *testing.T) {

	f64 := 73.73

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f64)
	if err != nil {
		t.Fatalf("float 64 to bytes err: %v", err)
	}

	f := bytesToFloat64(buf.Bytes())
	if err != nil {
		t.Fatalf("bytes to float err: %v", err)
	}

	if f != f64 {
		t.Fatalf("bytes to float converting result err: %v - %v", f64, f)
	}
	t.Logf("bytes to float converting OK: %v - %v", f64, f)
}

func Test_parseMessage_isValidReadMsg(t *testing.T) {

	testCases := []struct {
		name    string
		msg     readingMessage
		isValid bool
	}{
		// Positive
		{
			name: "valid message",
			msg: readingMessage{
				Temp:    0.0,
				Alt:     0.0,
				Lat:     0.0,
				Lon:     0.0,
				BattLev: 0.1,
			},
			isValid: true,
		},
		{
			name: "valid message 2",
			msg: readingMessage{
				Temp:    -300.0,
				Alt:     20000.0,
				Lat:     -90.0,
				Lon:     180.0,
				BattLev: 100.0,
			},
			isValid: true,
		},
		// Negative
		{
			name: "invalid message, temp out range",
			msg: readingMessage{
				Temp:    301.0,
				Alt:     0.0,
				Lat:     0.0,
				Lon:     0.0,
				BattLev: 0.1,
			},
			isValid: false,
		},
		{
			name: "invalid message, battery out range",
			msg: readingMessage{
				Temp:    300.0,
				Alt:     0.0,
				Lat:     0.0,
				Lon:     0.0,
				BattLev: 0.0,
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {

		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, tc.msg)
		if err != nil {
			t.Fatalf("message to bytes converting err: %v", err)
		}

		rm := readingMessage{}
		parseMessage(buf.Bytes(), &rm)
		ok := rm.isValid()
		if tc.isValid && !ok {
			t.Fatalf("%v: message %+v is invalid, parsed msg - %+v", tc.name, tc.msg, rm)
		} else if !tc.isValid && ok {
			t.Fatalf("%v: message %+v, negative test case shoud return not ok: parsed msg %+v", tc.name, tc.msg, rm)
		}
		if !reflect.DeepEqual(tc.msg, rm) {
			t.Fatalf("%v: orig message %+v and parsed %+v not equal", tc.name, tc.msg, rm)
		}
		t.Logf("%v: message %+v, test ok", tc.name, tc.msg)
	}
}
