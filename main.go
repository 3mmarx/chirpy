package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("New request!")
		http.NotFound(w, r)
	})

	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(s.ListenAndServe())
}
