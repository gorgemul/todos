package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorgemul/todos/pkg/server"
)

func TestGet(t *testing.T) {
	t.Run("get todo and return 200", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/", nil)
		response := httptest.NewRecorder()

		srv := server.New()

		srv.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("Want: %d, but got: %d", http.StatusOK, response.Code)
		}
	})
}
