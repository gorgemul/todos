package main

import (
	"log"
	"net/http"

	"github.com/gorgemul/todos/pkg/server"
)

func main() {
	srv := server.New()
	log.Println("listening port 8080")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
