package server

import (
	"context"
	"encoding/json"
	"fmt"
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

	result, err := json.MarshalIndent(todos, "", "    ")

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
	var newTodo types.NewTodo

	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		assertBadRequest(w, err)
		return
	}

	content := newTodo.Content

	if len(content) == 0 {
		http.Error(w, "content can't not be parsed!", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(context.Background(), "INSERT INTO todozz (content) VALUES ($1);", content)
	if err != nil {
		assertInternalErr(w, err)
		return
	}

	fmt.Fprintf(w, "Successfully add %q", content)
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	var updateTodo types.UpdateTodo

	err := json.NewDecoder(r.Body).Decode(&updateTodo)
	if err != nil {
		assertBadRequest(w, err)
		return
	}

	updateId := updateTodo.Id
	updateContent := updateTodo.Content

	if updateId <= 0 {
		http.Error(w, "invlid id", http.StatusBadRequest)
		return
	}

	if !todoExist(s.db, updateId) {
		http.Error(w, "update todo is not exist!", http.StatusBadRequest)
		return
	}

	if len(updateContent) == 0 {
		http.Error(w, "update content is empty!", http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(context.Background(), "UPDATE todozz SET content = $1 WHERE id = $2", updateContent, updateId)
	if err != nil {
		assertInternalErr(w, err)
		return
	}

	fmt.Fprintf(w, "Successfully update id: %d!!", updateId)
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from delete"))
}

func assertInternalErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	log.Println(err)
}

func assertBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
	log.Println(err)
}

func todoExist(db *pgx.Conn, id int) bool {
	err := db.QueryRow(context.Background(), "SELECT * FROM todozz WHERE id=$1", id).Scan(&id)
	return err != pgx.ErrNoRows
}
