package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorgemul/todos/pkg/server"
	"github.com/gorgemul/todos/types"
)

type stubStore struct {
	types.Todos
	newTodo types.NewTodo
}

func (s *stubStore) GetTodos() (types.Todos, error) {
	return s.Todos, nil
}

func (s *stubStore) PostTodo(content string) error {
	s.newTodo = types.NewTodo{Content: content}
	return nil
}

func (s *stubStore) UpdateTodo(id int, content string) error {
	return nil
}

func (s *stubStore) DeleteTodo(id int) error {
	return nil
}

func TestGet(t *testing.T) {
	t.Run("Get todos", func(t *testing.T) {
		want := types.Todos{
			{Id: 1, Content: "foo"},
			{Id: 2, Content: "bar"},
		}
		store := &stubStore{Todos: want}
		srv := server.New(store)

		request, err := newGetTodoRequest()
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		got := getTodosFromResponse(t, response)

		assertTodo(t, got, want)
	})
	t.Run("Get empty todos", func(t *testing.T) {
		want := types.Todos{}
		store := &stubStore{Todos: want}
		srv := server.New(store)

		request, err := newGetTodoRequest()
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		got := getTodosFromResponse(t, response)

		assertTodo(t, got, want)
	})
}

func TestPost(t *testing.T) {
	t.Run("Post valid todos", func(t *testing.T) {
		newTodo := types.NewTodo{Content: "legit todo content"}
		store := new(stubStore)
		srv := server.New(store)

		body := newRequestBody(t, newTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertTodo(t, store.newTodo, newTodo)
	})
	t.Run("Post invalid content", func(t *testing.T) {
		invalidNewTodo := map[string]string{
			"contnt": "something",
		}
		store := new(stubStore)
		srv := server.New(store)

		body := newRequestBody(t, invalidNewTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
	})
	t.Run("Post empty content", func(t *testing.T) {
		emptyNewTodo := types.NewTodo{Content: ""}
		store := new(stubStore)
		srv := server.New(store)

		body := newRequestBody(t, emptyNewTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
	})
}

func populateRequestBody(body io.Writer, v any) error {
	err := json.NewEncoder(body).Encode(v)
	if err != nil {
		return err
	}

	return nil
}

func getTodosFromResponse(t *testing.T, response *httptest.ResponseRecorder) types.Todos {
	var todos types.Todos
	err := json.NewDecoder(response.Body).Decode(&todos)
	if err != nil {
		t.Fatalf("problem get todos from response, %v", err)
	}
	return todos
}

func newGetTodoRequest() (*http.Request, error) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func newPostTodoRequest(body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest("POST", "/", body)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func newRequestBody(t *testing.T, v any) *bytes.Buffer {
	body := new(bytes.Buffer)
	err := populateRequestBody(body, v)
	if err != nil {
		t.Fatalf("problem populating reuqest body, %v", err)
	}

	return body
}

func assertTodo[T any](t testing.TB, got, want T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Want %v, but got %v", want, got)
	}
}

func assertErrMsg(t testing.TB, got, want string) {
	t.Helper()

	// NOTE: small workaounrd since http.Error() write w with Fprintln()
	comparedGot := strings.TrimSuffix(got, "\n")

	if want != comparedGot {
		t.Errorf("Want %v, but got %v", want, comparedGot)
	}
}

func assertNoErr(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("didn't expect an error but got one, %v", err)
	}
}
