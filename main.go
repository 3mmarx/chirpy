package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(".")))

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(s.ListenAndServe())
}
