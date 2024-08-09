package server

import (
	"net/http"
)

type Server struct {
	http.Handler
}

func New() *Server {
	srv := new(Server)
	mux := http.NewServeMux()

	mux.Handle("GET /", http.HandlerFunc(srv.getHandler))
	mux.Handle("POST /", http.HandlerFunc(srv.postHandler))
	mux.Handle("PUT /update", http.HandlerFunc(srv.putHandler))
	mux.Handle("DELETE /delete", http.HandlerFunc(srv.deleteHandler))

	srv.Handler = mux
	return srv
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from get"))
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
