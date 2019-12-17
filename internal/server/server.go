// Package server implements a concurrent logging server for thermometers.
package server

import (
	"log"
	"net"
)

// Config server configs
type Config struct {
	Addr string
}

// Server implement logging server of thermometers.
type Server struct {
	conf Config
}

// New init new Server.
func New(conf Config) *Server {
	s := &Server{conf: conf}
	return s
}

// Start starts new server.
func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		log.Printf("new listenner, addr - %v, err: %v", s.conf.Addr, err)
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept new connect, err: %v", err)
			continue
		}
		log.Printf("new conn accepted: laddr - %v, raddr - %v", conn.LocalAddr(), conn.RemoteAddr())

		d := newDevice(devConfig{}, conn)
		go d.run()
	}

}
