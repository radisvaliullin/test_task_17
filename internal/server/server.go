// Package server implements a concurrent logging server for thermometers.
package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config server configs
type Config struct {
	// server address
	Addr string
	// http server address
	HTTPAddr string

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

	// run server
	s.wg.Add(1)
	go func() {
		if err := s.run(); err != nil {
			s.errs <- err
		}
	}()

	// run HTTP server
	go func() {
		if err := s.startHTTPServer(); err != nil {
			log.Printf("http server down: %v", err)
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

// http server
func (s *Server) startHTTPServer() error {

	mux := http.NewServeMux()
	mux.HandleFunc("/stats", s.stats)
	mux.HandleFunc("/readings/", s.readings)
	mux.HandleFunc("/status/", s.status)

	return http.ListenAndServe(s.conf.HTTPAddr, mux)
}

// not implemented
func (s *Server) stats(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if _, err := w.Write([]byte("Not Implemented")); err != nil {
		log.Printf("http server: write err: %v", err)
	}
}

// return last Reading of device by IMEI
func (s *Server) readings(w http.ResponseWriter, req *http.Request) {
	imei := strings.TrimPrefix(req.URL.Path, "/readings/")
	if _, err := strconv.ParseInt(imei, 10, 64); err != nil {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("404 Not Found")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}
	if len(imei) != 15 {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("500 Wrong IMEI length")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}

	// request last reading
	drs := deviceReadingStatus{
		deviceStatus: deviceStatus{
			IMEI: imei,
		},
	}
	dreq, ok := s.devStor.ok(imei)
	if ok {
		dresp := make(chan deviceReadingStatus, 1)
		dreq <- dresp
		resp, ok := <-dresp
		if ok {
			drs.Status = "online"
			drs.Reading = resp.Reading
			drs.Time = resp.Time
		} else {
			drs.Status = "offline"
		}
	} else {
		drs.Status = "offline"
	}

	// response
	out, err := json.Marshal(&drs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("500 Internal Server Error")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(out); err != nil {
		log.Printf("http server: write err: %v", err)
	}
}

// return device status by IMEI
func (s *Server) status(w http.ResponseWriter, req *http.Request) {
	// get IMEI from Path
	imei := strings.TrimPrefix(req.URL.Path, "/status/")
	if _, err := strconv.ParseInt(imei, 10, 64); err != nil {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("404 Not Found")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}
	if len(imei) != 15 {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("500 Wrong IMEI length")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}

	// response status
	sts := deviceStatus{
		IMEI: imei,
	}
	if _, ok := s.devStor.ok(imei); ok {
		sts.Status = "online"
	} else {
		sts.Status = "offline"
	}

	// response
	out, err := json.Marshal(&sts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("500 Internal Server Error")); err != nil {
			log.Printf("http server: write err: %v", err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(out); err != nil {
		log.Printf("http server: write err: %v", err)
	}
}
