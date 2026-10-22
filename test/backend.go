package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// go run test/backend.go -port 8081
// go run test/backend.go -port 8081

func main() {
	port := flag.Int("port", 8081, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Backend %d] Received request: %s %s", *port, r.Method, r.URL.Path)
		fmt.Fprintf(w, "Hello from Backend (Port %d)\n", *port)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
