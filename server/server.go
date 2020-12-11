package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keyan/simpledb/database"
	"github.com/keyan/simpledb/rpc"
)

const (
	ReadTimeout  = 1 * time.Second
	WriteTimeout = 1 * time.Second
)

// Run creates a new database (which reloads data from disk), then starts a HTTP
// server to handle client database operation requests.
func Run(serverAddr string) {
	db := database.New()

	s := &http.Server{
		Addr:    serverAddr,
		Handler: nil, // Default handler
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := rpc.DecodeMsg(r.Body)
		switch msg.Op {
		case rpc.Get:
			val, err := db.Get(msg.Key)
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			} else {
				fmt.Fprintf(w, string(val))
			}
		case rpc.Set:
			db.Set(msg.Key, msg.Value)
		case rpc.Delete:
			db.Delete(msg.Key)
		default:
			fmt.Fprintf(w, "Unknown operation type")
		}
	})

	log.Fatal(s.ListenAndServe())
}
