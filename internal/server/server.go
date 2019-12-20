// Package server implements a concurrent logging server for thermometers.
package server

import (
	"log"
	"net"
	"sync"
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

	// listener
	ln net.Listener

	// wg
	wg sync.WaitGroup

	// server error
	errs chan error

	//
	devStor *devStorage
}

// New inits new Server.
func New(conf Config, olg *log.Logger) *Server {
	s := &Server{
		conf:    conf,
		outLog:  olg,
		errs:    make(chan error, 1),
		devStor: newDevStorage(),
	}
	return s
}

// Start starts new server.
func (s *Server) Start() error {

	log.Print("server listener starting ", s.conf.Addr)
	ln, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		log.Printf("new listener, addr - %v, err: %v", s.conf.Addr, err)
		return err
	}
	s.ln = ln

	s.wg.Add(1)
	go func() {
		if err := s.run(); err != nil {
			s.errs <- err
		}
	}()

	return nil
}

// Stop stops server
func (s *Server) Stop() {
	if err := s.ln.Close(); err != nil {
		log.Printf("server, listenner close err: %v", err)
	}
}

// Wait waits server stoping (blocking)
func (s *Server) Wait() {
	s.wg.Wait()
}

// Error server error
func (s *Server) Error() <-chan error {
	return s.errs
}

func (s *Server) run() error {
	defer s.wg.Done()
	// stop handler signal
	stop := make(chan struct{}, 1)
	defer func() {
		stop <- struct{}{}
	}()

	for {
		log.Print("server, waiting new client")
		conn, err := s.ln.Accept()
		if err != nil {
			log.Printf("accept new connect, err: %v", err)
			break
		}
		log.Printf("new conn accepted: laddr - %v, raddr - %v", conn.LocalAddr(), conn.RemoteAddr())

		// connection (device) handler responsible for close connection
		s.wg.Add(1)
		d := newDevice(
			devConfig{loginDeadline: s.conf.LoginDeadline, messageDeadline: s.conf.MsgDeadline},
			conn, s.outLog, &s.wg, stop, s.devStor,
		)
		go d.run()
	}

	return nil
}
