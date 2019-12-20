package main

import (
	"log"
	"time"

	"github.com/radisvaliullin/test_task_17/internal/server"
)

func main() {

	// simple client
	conf := server.TestClientConfig{
		SrvAddr:        ":1337",
		IMEI:           [15]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8},
		PeriodDuration: time.Millisecond * 500,
	}
	cln := server.NewTestClient(conf)
	cln.Start()
	cln.Wait()
	select {
	case err := <-cln.Error():
		log.Printf("client stoped, err: %v", err)
	default:
	}
}
