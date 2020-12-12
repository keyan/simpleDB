package main

import (
	"flag"
	"math/rand"
	"os/exec"
	"time"

	"github.com/keyan/simpledb/client"
	"github.com/keyan/simpledb/server"
)

const (
	serverAddr = "localhost:8080"

	simClients          = 10
	simClientPingRateMs = 100
)

func main() {
	clientMode := flag.Bool("client", false, "Run simpleDB in client CLI mode")
	simMode := flag.Bool(
		"simulation",
		false,
		"Run simpleDB in simulation mode to test concurrency and fault-tolerance",
	)
	flag.Parse()

	switch {
	case *simMode:
		simulation()
	case *clientMode:
		client.SetServerAddr(serverAddr)
		client.Run()
	default:
		server.Run(serverAddr)
	}
}

func simulation() {
	cmd := exec.Command("./simpledb")
	// cmd.Start()

	client.SetServerAddr(serverAddr)
	for i := 0; i < simClients; i++ {
		client.RunSimClient(simClientPingRateMs)
	}

	for {
		val := rand.Intn(100)
		if val > 100 {
			cmd.Process.Kill()
		}

		time.Sleep(1 * time.Second)
	}
}
