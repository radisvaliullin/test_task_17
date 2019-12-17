package server

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	imeiLength = 15
	msgLength  = 40
)

type devConfig struct {
	loginDeadline   time.Duration
	messageDeadline time.Duration
}

type device struct {
	conf devConfig
	conn net.Conn
}

func newDevice(conf devConfig, conn net.Conn) *device {
	return &device{conf: conf, conn: conn}
}

func (d *device) run() error {
	// close connection
	defer d.conn.Close()

	// device login
	// imei read
	imei := make([]byte, 15)
	d.conn.SetReadDeadline(time.Now().Add(d.conf.loginDeadline))
	_, err := io.ReadFull(d.conn, imei)
	if err != nil {
		log.Printf("device: read imei err: %v", err)
		return err
	}
	log.Printf("device: got imei: %v", string(imei))

	// read messages in cycle
	msg := make([]byte, 40)
	for {

		// read message
		d.conn.SetReadDeadline(time.Now().Add(d.conf.messageDeadline))
		_, err := io.ReadFull(d.conn, msg)
		if err != nil {
			log.Printf("device: read message err: %v", err)
			return err
		}
	}
}
