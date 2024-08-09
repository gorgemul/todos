package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorgemul/todos/types"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	db *pgx.Conn
	http.Handler
}

func New(db *pgx.Conn) *Server {
	srv := new(Server)

	srv.db = db
	mux := http.NewServeMux()

	mux.Handle("GET /", http.HandlerFunc(srv.getHandler))
	mux.Handle("POST /", http.HandlerFunc(srv.postHandler))
	mux.Handle("PUT /update", http.HandlerFunc(srv.putHandler))
	mux.Handle("DELETE /delete", http.HandlerFunc(srv.deleteHandler))

	srv.Handler = mux
	return srv
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(context.Background(), "SELECT * FROM todozz")

	if err != nil {
		assertInternalErr(w, err)
		return
	}

	var todos []types.Todo

	for rows.Next() {
		var todo types.Todo
		if err := rows.Scan(&todo.Id, &todo.Content, &todo.CreatedAt); err != nil {
			assertInternalErr(w, err)
			return
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		assertInternalErr(w, err)
		return
	}

	result, err := json.Marshal(todos)

	if err != nil {
		assertInternalErr(w, err)
		return
	}

	_, err = w.Write(result)

	if err != nil {
		assertInternalErr(w, err)
		return
	}
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from post"))
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from put"))
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from delete"))
}

func assertInternalErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	log.Println(err)
}
