package main

import (
	"flag"
	"fmt"
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
		runSimulation()
	case *clientMode:
		client.SetServerAddr(serverAddr)
		client.Run()
	default:
		server.Run(serverAddr)
	}
}

// startServerProcess starts the DB server in a separate forked process.
func startServerProcess() *exec.Cmd {
	fmt.Println("starting server process")
	cmd := exec.Command("./simpledb")
	cmd.Start()
	return cmd
}

// runSimulation starts a DB server in new process, starts a bunch of simulated
// clients that make random requests, and occasionally restarts the DB server.
func runSimulation() {
	cmd := startServerProcess()

	client.SetServerAddr(serverAddr)
	for i := 0; i < simClients; i++ {
		client.RunSimClient(simClientPingRateMs)
	}

	for {
		val := rand.Intn(100)
		if val > 90 {
			fmt.Println("forcibly killing server process")
			cmd.Process.Kill()
			cmd = startServerProcess()
		}

		time.Sleep(500 * time.Millisecond)
	}
}
