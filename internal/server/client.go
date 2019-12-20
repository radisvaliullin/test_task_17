package server

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"time"
)

// TestClientConfig configs of TestClient
type TestClientConfig struct {
	SrvAddr string
	IMEI    [15]byte

	// message resend period duration
	PeriodDuration time.Duration
}

// TestClient server's test client. Connect and send periodically test messages.
type TestClient struct {
	conf TestClientConfig

	// error chan
	errs chan error

	// stop signal
	stop chan struct{}
	wg   sync.WaitGroup
}

// NewTestClient inits new TestClient
func NewTestClient(conf TestClientConfig) *TestClient {
	tc := &TestClient{
		conf: conf,
		errs: make(chan error, 1),
		stop: make(chan struct{}, 1),
	}
	return tc
}

// Start starts client
func (c *TestClient) Start() {

	c.wg.Add(1)
	go func() {
		if err := c.run(); err != nil {
			c.errs <- err
		}
	}()
}

// Stop stops client
func (c *TestClient) Stop() {
	c.stop <- struct{}{}
}

// Wait Client stopped (blocked)
func (c *TestClient) Wait() {
	c.wg.Wait()
}

// Error returns error chan
func (c *TestClient) Error() <-chan error {
	return c.errs
}

func (c *TestClient) run() error {
	defer c.wg.Done()

	// connect
	conn, err := net.Dial("tcp", c.conf.SrvAddr)
	if err != nil {
		log.Printf("client conn err: %v", err)
		return err
	}
	defer conn.Close()
	log.Printf("client addr - %v, connected to server - %v", conn.LocalAddr(), conn.RemoteAddr())

	// send login (IMEI)
	_, err = conn.Write(c.conf.IMEI[:])
	if err != nil {
		log.Printf("client, addr = %v, send imei err: %v", conn.LocalAddr(), err)
		return err
	}
	log.Printf("client, addr - %v, imei sent - %v", conn.LocalAddr(), c.conf.IMEI)

	for {
		select {
		case <-c.stop:
			return nil
		default:

			// send message
			msg := readingMessage{
				Temp: 0.0,
			}
			// message to bytes
			var buf bytes.Buffer
			err := binary.Write(&buf, binary.BigEndian, msg)
			if err != nil {
				log.Printf("client, imei - %v, message to bytes converting err: %v", c.conf.IMEI, err)
				return err
			}

			// send
			_, err = conn.Write(buf.Bytes())
			if err != nil {
				log.Printf("client, imei - %v, send message err: %v", c.conf.IMEI, err)
				return err
			}
			log.Printf("client, imei - %v, message %+v sent", c.conf.IMEI, msg)

			// sleep
			time.Sleep(c.conf.PeriodDuration)
		}
	}
}
