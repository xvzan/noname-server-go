package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "WebSocket server port")
	flag.Parse()
	http.HandleFunc("/", handleWebSocket)
	// log.Printf("WebSocket server started on :%s", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
