package client

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz"
	strLen      = 3
)

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RunSimClient(pingRateMs int) {
	fmt.Println("starting simulation client...")
	go func() {
		for {
			time.Sleep(time.Duration(pingRateMs) * time.Millisecond)

			key := randStringBytes(strLen)

			switch roll := rand.Intn(10); {
			case roll < 3:
				doGet(key)
			case roll < 7:
				value := randStringBytes(strLen)
				doSet(key, value)
			default:
				doDelete(key)
			}

		}
	}()
}
