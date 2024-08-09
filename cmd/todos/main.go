package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorgemul/todos/pkg/db"
	"github.com/gorgemul/todos/pkg/server"
)

func main() {
	db, err := db.New()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close(context.Background())

	srv := server.New(db)
	log.Println("listening port 8080")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
