package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorgemul/todos/pkg/db"
	"github.com/gorgemul/todos/types"
)

const (
	InvalidContentErrMsg = "Invalid content!"
	InvalidIdErrMsg      = "Invalid id!"
)

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
	mux.Handle("PUT /update", http.HandlerFunc(srv.putHandler))
	// mux.Handle("DELETE /delete/{id}", http.HandlerFunc(srv.deleteHandler))

	srv.Handler = mux
	return srv
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := s.store.GetTodos()
	if err != nil {
		s.logAndResponse(w, err, http.StatusInternalServerError)
		return
	}

	if err := s.responseInJSON(w, todos); err != nil {
		s.logAndResponse(w, err, http.StatusInternalServerError)
		return
	}
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	content, err := s.extractContentFromRequest(r)
	if err != nil {
		s.logAndResponse(w, err, http.StatusInternalServerError)
		return
	}

	if !s.validContent(content) {
		s.logAndResponse(w, errors.New(InvalidContentErrMsg), http.StatusBadRequest)
		return
	}

	if err := s.store.PostTodo(content); err != nil {
		s.logAndResponse(w, err, http.StatusInternalServerError)
		return
	}

	s.dbExecuteSuccess(w, "add new todo")
}

func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	id, content, err := s.extractIdAndContentFromRequest(r)
	if err != nil {
		s.logAndResponse(w, err, http.StatusInternalServerError)
		return
	}

	var validParamsErr error

	switch {
	case !s.validId(id):
		validParamsErr = errors.New(InvalidIdErrMsg)
	case !s.validContent(content):
		validParamsErr = errors.New(InvalidContentErrMsg)
	}

	if validParamsErr != nil {
		s.logAndResponse(w, validParamsErr, http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateTodo(id, content); err != nil {
		switch err {
		case db.UpdatedIdNotExistErr:
			s.logAndResponse(w, err, http.StatusBadRequest)
		default:
			s.logAndResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	s.dbExecuteSuccess(w, "update todo")
}

func (s *Server) logAndResponse(w http.ResponseWriter, err error, code int) {
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

	return newTodo.Content, nil
}

func (s *Server) extractIdAndContentFromRequest(r *http.Request) (int, string, error) {
	var updateTodo types.UpdateTodo
	err := json.NewDecoder(r.Body).Decode(&updateTodo)
	if err != nil {
		return 0, "", err
	}

	return updateTodo.Id, updateTodo.Content, nil
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

func (s *Server) validId(id int) bool {
	return id > 0
}

func (s *Server) validContent(content string) bool {
	return len(content) > 0
}
