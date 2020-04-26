package main

import (
	"log"
	"net/http"

	"./chat"
)

func main() {
	server := chat.NewServer("/entry")
	go server.Listen()
	http.Handle("/", http.FileServer(http.Dir("webroot")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
