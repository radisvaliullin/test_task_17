package main

import "github.com/radisvaliullin/test_task_17/internal/server"

func main() {

	s := server.New(server.Config{Addr: ":1337"})

	s.Start()
}
