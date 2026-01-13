package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := ":3000"
	log.Printf("Сервер запущен на http://localhost:3000\n")
	log.Fatal(http.ListenAndServe(port, nil))
}
