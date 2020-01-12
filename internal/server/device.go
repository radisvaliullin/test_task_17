package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"
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

// device handle connection with new devices
type device struct {
	conf devConfig

	// device connection
	conn  net.Conn
	raddr string
	imei  string

	// dev stor
	devStor *devStorage

	// outlog for logging Reading messages
	outLog *log.Logger

	// wg
	wg *sync.WaitGroup
	// server stop sig
	srvStop chan struct{}
}

// inits new device
func newDevice(
	conf devConfig, conn net.Conn, olg *log.Logger, wg *sync.WaitGroup, stop chan struct{}, ds *devStorage,
) *device {
	d := &device{
		conf:    conf,
		conn:    conn,
		raddr:   conn.RemoteAddr().String(),
		outLog:  olg,
		wg:      wg,
		srvStop: stop,
		devStor: ds,
	}
	return d
}

// run handle new connection (responsible to close connection)
func (d *device) run() error {
	// stopped signal
	stopped := make(chan struct{}, 1)
	//
	defer d.wg.Done()
	// close connection
	defer func() {
		stopped <- struct{}{}
		if err := d.conn.Close(); err != nil {
			log.Printf("device conn close err: %v", err)
		}
	}()
	// server stop handler
	go func() {
		select {
		case <-stopped:
		case <-d.srvStop:
			if err := d.conn.Close(); err != nil {
				log.Printf("device conn close err: %v", err)
			}
			d.srvStop <- struct{}{}
		}
	}()

	// device login
	// imei read
	log.Printf("device logging, raddr - %v", d.raddr)
	imei := make([]byte, imeiLength)
	// set login deadline
	d.conn.SetReadDeadline(time.Now().Add(d.conf.loginDeadline))
	// read imei
	_, err := io.ReadFull(d.conn, imei)
	if err != nil {
		log.Printf("device, raddr - %v, read imei err: %v", d.raddr, err)
		return err
	}
	// parse imei
	d.imei, err = validParseIMEI(imei)
	if err != nil {
		log.Printf("device raddr - %v, imei validate err: %v", d.raddr, err)
		return err
	}
	// register device by imei
	dreq := make(devReq, 1)
	if ok := d.devStor.setIfNot(d.imei, dreq); !ok {
		log.Printf("device, raddr - %v, device with imei - %v yet registered", d.raddr, d.imei)
		return fmt.Errorf("device with imei %v yet registered", d.imei)
	}
	// unregister when connection closed
	defer func() {
		d.devStor.delete(d.imei)
	}()
	log.Printf("device logged, raddr - %v, imei %v", d.raddr, d.imei)

	// read messages in cycle
	msg := make([]byte, 40)
	rm := readingMessage{}
	for {

		// read message
		d.conn.SetReadDeadline(time.Now().Add(d.conf.messageDeadline))
		_, err := io.ReadFull(d.conn, msg)
		if err != nil {
			log.Printf("device, imei - %v, read message err: %v", d.imei, err)
			return err
		}
		now := time.Now().UnixNano()

		// parse message
		parseMessage(msg, &rm)
		log.Printf("device, imei - %v, read message %+v", d.imei, rm)

		// if valid, logging Reading message to stdout
		if rm.isValid() {
			reading := fmt.Sprintf("%v,%s,%f,%f,%f,%f,%f\n", now, d.imei, rm.Temp, rm.Alt, rm.Lat, rm.Lon, rm.BattLev)
			d.outLog.Print(reading)

			// response to request last reading
			select {
			case dresp := <-dreq:
				dresp <- deviceReadingStatus{Reading: rm, Time: now}
				close(dresp)
			default:
			}
		} else {
			log.Printf("device, imei %v, invalid reading message %+v", d.imei, rm)
		}
	}
}

// validate imei (Luhn algorithm), parse to string
func validParseIMEI(imei []byte) (string, error) {
	if len(imei) != imeiLength {
		return "", errors.New("imei wrong length")
	}
	// imei chars
	imeiChr := make([]byte, imeiLength)
	// result checksum
	var check byte
	for i, b := range imei[:imeiLength-1] {
		// each byte should be decimal (0-9)
		if b > 9 {
			return "", errors.New("imei should be decimal number (0-9)")
		}
		if i%2 > 0 {
			b = b * 2
			if b > 9 {
				b = b/10 + b%10
			}
		}
		check += b
		// add 48 to get ASCII code of current decimal number
		imeiChr[i] = imei[i] + 48
	}
	check = check % 10
	if check != 0 {
		check = 10 - check
	}
	if check != imei[imeiLength-1] {
		return "", errors.New("imei wrong check")
	}
	// add 48 to get ASCII code of last decimal number
	imeiChr[imeiLength-1] = imei[imeiLength-1] + 48
	return string(imeiChr), nil
}

// parse Reading message
func parseMessage(msg []byte, rm *readingMessage) {
	// panic if len less then message length
	_ = msg[msgLength-1]
	//
	rm.Temp = bytesToFloat64(msg[:8])
	rm.Alt = bytesToFloat64(msg[8:16])
	rm.Lat = bytesToFloat64(msg[16:24])
	rm.Lon = bytesToFloat64(msg[24:32])
	rm.BattLev = bytesToFloat64(msg[32:])
}

// covert 8-bytes to float64
func bytesToFloat64(b []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}
