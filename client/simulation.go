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
	go func() {
		for {
			time.Sleep(time.Duration(pingRateMs) * time.Millisecond)

			key := randStringBytes(strLen)

			switch roll := rand.Intn(10); {
			case roll < 3:
				resp := doGet(key)
				fmt.Printf("Get op, key: %s, val: %s\n", key, resp)
			case roll < 8:
				value := randStringBytes(strLen)
				doSet(key, value)
				fmt.Printf("Set op, key: %s, val: %s\n", key, value)
			default:
				doDelete(key)
			}

		}
	}()
}
