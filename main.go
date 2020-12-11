package main

import (
	"flag"

	"github.com/keyan/simpledb/client"
	"github.com/keyan/simpledb/server"
)

const (
	serverAddr = "localhost:8080"
)

func main() {
	clientMode := flag.Bool("client", false, "Run simpleDB in client CLI mode")
	flag.Parse()

	if *clientMode {
		client.Run(serverAddr)
	} else {
		server.Run(serverAddr)
	}
}
