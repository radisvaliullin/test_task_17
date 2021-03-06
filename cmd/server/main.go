package main

import (
	"log"
	"os"
	"time"

	"github.com/radisvaliullin/test_task_17/internal/server"
)

func main() {

	// config init server
	log.Print("server init")

	// stdout logger (for logging server reading messages)
	outLog := log.New(os.Stdout, "", 0)

	// new server init
	s := server.New(server.Config{
		Addr: ":1337", HTTPAddr: ":1338", LoginDeadline: time.Second, MsgDeadline: time.Second * 2},
		outLog,
	)

	log.Print("server starting")
	err := s.Start()
	if err != nil {
		log.Fatalf("server starting err: %v", err)
	}
	s.Wait()
}
