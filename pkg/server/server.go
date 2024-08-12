package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorgemul/todos/types"
	"github.com/jackc/pgx/v5"
)

const InvalidContentErrMsg = "Invalid content!"

type TodoStore interface {
	GetTodos() (types.Todos, error)
	PostTodo(content string) error
	UpdateTodo(id int, content string) error
	DeleteTodo(id int) error
}

type Server struct {
	store TodoStore
	http.Handler
}

func New(store TodoStore) *Server {
	srv := new(Server)

	srv.store = store
	mux := http.NewServeMux()

	mux.Handle("GET /", http.HandlerFunc(srv.getHandler))
	mux.Handle("POST /", http.HandlerFunc(srv.postHandler))
	// mux.Handle("PUT /update", http.HandlerFunc(srv.putHandler))
	// mux.Handle("DELETE /delete/{id}", http.HandlerFunc(srv.deleteHandler))

	srv.Handler = mux
	return srv
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := s.store.GetTodos()
	if err != nil {
		s.logAndResponseWithStatus(w, err, http.StatusInternalServerError)
		return
	}

	err = s.responseInJSON(w, todos)
	if err != nil {
		s.logAndResponseWithStatus(w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	content, err := s.extractContentFromRequest(r)
	if err != nil {
		s.logAndResponseWithStatus(w, err, http.StatusInternalServerError)
		return
	}

	err = s.store.PostTodo(content)
	if err != nil {
		s.logAndResponseWithStatus(w, err, http.StatusInternalServerError)
		return
	}

	s.dbExecuteSuccess(w, "add new todo")
}

func (s *Server) logAndResponseWithStatus(w http.ResponseWriter, err error, code int) {
	errMsg := err.Error()
	log.Println(errMsg)
	http.Error(w, errMsg, code)
}

func (s *Server) extractContentFromRequest(r *http.Request) (string, error) {
	var newTodo types.NewTodo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		return "", err
	}

	if len(newTodo.Content) == 0 {
		return "", errors.New(InvalidContentErrMsg)
	}
	return newTodo.Content, nil
}

func (s *Server) responseInJSON(w http.ResponseWriter, v any) error {
	byte, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return fmt.Errorf("problem marshal indent format JSON, %v", err)
	}

	_, err = w.Write(byte)
	if err != nil {
		return fmt.Errorf("problem writing json response, %v", err)
	}

	return nil
}

func (s *Server) dbExecuteSuccess(w http.ResponseWriter, msg string) {
	fmt.Fprintf(w, "Successfully %s!!!\n", msg)
}

func (s *Server) todoExist(db *pgx.Conn, id int) bool {
	err := db.QueryRow(context.Background(), "SELECT * FROM todozz WHERE id=$1", id).Scan(&id)
	return err != pgx.ErrNoRows
}
