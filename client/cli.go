package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/keyan/simpledb/rpc"
)

const (
	// This isn't an official MIME type, so it doesn't get used by the server.
	contentType = "application/x-gob"

	promptColor = "\033[1;36m%s\033[0m"
)

var serverAddr string

func SetServerAddr(addr string) {
	serverAddr = "http://" + addr
}

func Run() {
	fmt.Printf("-- SimpleDB CLI --\n\n")
	printHelp()
	for {
		fmt.Printf(promptColor, "simpleDB > ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		rawInput := scanner.Text()

		sepInput := strings.Split(rawInput, " ")
		switch numArgs := len(sepInput); {
		case numArgs == 0:
			fmt.Println("Must provide one of the available command options")
		case numArgs >= 4:
			fmt.Println("Too many options provided")
		}

		switch sepInput[0] {
		case "help":
			printHelp()
		case "exit":
			os.Exit(0)
		case "get":
			if len(sepInput) != 2 {
				fmt.Println("Invalid number of options provided")
			} else {
				resp := doGet(sepInput[1])
				fmt.Printf("%s\n", resp)
			}
		case "set":
			if len(sepInput) != 3 {
				fmt.Println("Invalid number of options provided")
			} else {
				doSet(sepInput[1], sepInput[2])
			}
		case "delete":
			if len(sepInput) != 2 {
				fmt.Println("Invalid number of options provided")
			} else {
				doDelete(sepInput[1])
			}
		default:
			fmt.Println("Invalid command provided")
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Printf("%v%v \n", strings.Repeat(" ", 5), "help")
	fmt.Printf("%v%v \n", strings.Repeat(" ", 5), "exit or CTRL+C")
	fmt.Printf("%v%v \n", strings.Repeat(" ", 5), "get <key>")
	fmt.Printf("%v%v \n", strings.Repeat(" ", 5), "set <key> <value>")
	fmt.Printf("%v%v \n", strings.Repeat(" ", 5), "delete <key>")
}

func doGet(key string) string {
	msg := rpc.NewGetMsg(key)
	return callServer(msg)
}

func doSet(key string, val string) {
	msg := rpc.NewSetMsg(key, rpc.ValueType(val))
	callServer(msg)
}

func doDelete(key string) {
	msg := rpc.NewDeleteMsg(key)
	callServer(msg)
}

func callServer(msg *bytes.Buffer) string {
	resp, err := http.Post(serverAddr, contentType, msg)
	if err != nil {
		fmt.Printf("Could not issue command, err: %v\n", err)
		return ""
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}
