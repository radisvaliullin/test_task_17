// Package server implements a concurrent logging server for thermometers.
package server

import (
	"log"
	"net"
	"time"
)

// Config server configs
type Config struct {
	// server address
	Addr string

	// client message read timeouts
	LoginDeadline time.Duration
	MsgDeadline   time.Duration
}

// Server implements logging server of thermometers.
type Server struct {
	conf Config

	// std out logger for loggin Reading message
	outLog *log.Logger
}

// New inits new Server.
func New(conf Config, olg *log.Logger) *Server {
	s := &Server{conf: conf, outLog: olg}
	return s
}

// Start starts new server.
func (s *Server) Start() error {

	log.Print("server listener starting")
	ln, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		log.Printf("new listener, addr - %v, err: %v", s.conf.Addr, err)
		return err
	}

	for {
		log.Print("server, waiting new client")
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept new connect, err: %v", err)
			continue
		}
		log.Printf("new conn accepted: laddr - %v, raddr - %v", conn.LocalAddr(), conn.RemoteAddr())

		// connection (device) handler responsible for close connection
		d := newDevice(devConfig{loginDeadline: s.conf.LoginDeadline, messageDeadline: s.conf.MsgDeadline}, conn, s.outLog)
		go d.run()
	}

}
