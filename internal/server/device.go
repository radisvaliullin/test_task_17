package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

// device handle connection with new devices
type device struct {
	conf devConfig

	// device connection
	conn  net.Conn
	raddr string
	imei  string

	// outlog for logging Reading messages
	outLog *log.Logger
}

// inits new device
func newDevice(conf devConfig, conn net.Conn, olg *log.Logger) *device {
	return &device{conf: conf, conn: conn, raddr: conn.RemoteAddr().String(), outLog: olg}
}

// run handle new connection (responsible to close connection)
func (d *device) run() error {
	// close connection
	defer d.conn.Close()

	// device login
	// imei read
	log.Printf("device logging, raddr - %v", d.raddr)
	imei := make([]byte, imeiLength)
	// set login deadline
	d.conn.SetReadDeadline(time.Now().Add(d.conf.loginDeadline))
	_, err := io.ReadFull(d.conn, imei)
	if err != nil {
		log.Printf("device, raddr - %v, read imei err: %v", d.raddr, err)
		return err
	}
	d.imei, err = validParseIMEI(imei)
	if err != nil {
		log.Printf("device raddr - %v, imei validate err: %v", d.raddr, err)
		return err
	}
	log.Printf("device logged, raddr - %v, imei %v", d.raddr, d.imei)

	// read messages in cycle
	msg := make([]byte, 40)
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
		r, err := parseMessage(msg)
		if err != nil {
			log.Printf("device, imei - %v, message parse err: %v", d.imei, err)
			return err
		}
		log.Printf("device, imei - %v, read message %+v", d.imei, r)

		// if valid, logging Reading message to stdout
		if err := isValidReadMsg(r); err == nil {
			reading := fmt.Sprintf("%v,%s,%f,%f,%f,%f,%f\n", now, d.imei, r.Temp, r.Alt, r.Lat, r.Lon, r.BattLev)
			d.outLog.Print(reading)
		} else {
			log.Printf("device, imei %v, invalid reading message %+v", d.imei, r)
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
func parseMessage(msg []byte) (readingMessage, error) {

	rm := readingMessage{}

	if len(msg) != msgLength {
		return rm, errors.New("message wrong length")
	}

	var err error
	// temp
	rm.Temp, err = bytesToFloat64(msg[:8])
	if err != nil {
		return rm, err
	}
	// alt
	rm.Alt, err = bytesToFloat64(msg[8:16])
	if err != nil {
		return rm, err
	}
	// lat
	rm.Lat, err = bytesToFloat64(msg[16:24])
	if err != nil {
		return rm, err
	}
	// lon
	rm.Lon, err = bytesToFloat64(msg[24:32])
	if err != nil {
		return rm, err
	}
	// battery level
	rm.BattLev, err = bytesToFloat64(msg[32:])
	if err != nil {
		return rm, err
	}

	return rm, nil
}

// covert 8-bytes to float64
func bytesToFloat64(b []byte) (float64, error) {

	if len(b) != 8 {
		return 0.0, errors.New("bytes length should be 8")
	}
	ui64 := (uint64(b[0]) << 56) | (uint64(b[1]) << 48) | (uint64(b[2]) << 40) | (uint64(b[3]) << 32)
	ui64 |= (uint64(b[4]) << 24) | (uint64(b[5]) << 16) | (uint64(b[6]) << 8) | uint64(b[7])

	return math.Float64frombits(ui64), nil
}

// validate Reading message
func isValidReadMsg(rm readingMessage) error {
	// temp
	if rm.Temp < -300.0 || rm.Temp > 300.0 {
		return errors.New("message, temperatue is out range [-300, 300]")
	}
	// alt
	if rm.Alt < -20000.0 || rm.Alt > 20000.0 {
		return errors.New("message, altitude is out range [-300, 300]")
	}
	// lat
	if rm.Lat < -90.0 || rm.Lat > 90.0 {
		return errors.New("message, altitude is out range [-300, 300]")
	}
	// lon
	if rm.Lon < -180.0 || rm.Lon > 180.0 {
		return errors.New("message, longitude is out range [-300, 300]")
	}
	// battery level
	if rm.BattLev < 0.0 || rm.BattLev > 100.0 {
		return errors.New("message, battery level is out range [-300, 300]")
	}
	return nil
}
