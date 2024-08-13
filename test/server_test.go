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
	"time"

	"github.com/gorgemul/todos/pkg/db"
	"github.com/gorgemul/todos/pkg/server"
	"github.com/gorgemul/todos/types"
)

type stubStore struct {
	types.Todos
	newTodo types.NewTodo
}

var (
	dummyStore = new(stubStore)
	dummyTime  = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	dummyTodos = types.Todos{
		{Id: 1, Content: "foo", CreatedAt: dummyTime},
		{Id: 2, Content: "bar", CreatedAt: dummyTime},
	}
)

func (s *stubStore) GetTodos() (types.Todos, error) {
	return s.Todos, nil
}

func (s *stubStore) PostTodo(content string) error {
	s.newTodo = types.NewTodo{Content: content}
	return nil
}

func (s *stubStore) UpdateTodo(id int, content string) error {
	for i, todo := range s.Todos {
		if todo.Id == id {
			s.Todos[i].Content = content
			return nil
		}
	}
	return db.UpdatedIdNotExistErr
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
		assertStatus(t, response.Code, http.StatusOK)
	})
}

func TestPost(t *testing.T) {
	t.Run("Post happy path", func(t *testing.T) {
		newTodo := types.NewTodo{Content: "legit todo content"}
		srv := server.New(dummyStore)

		body := newRequestBody(t, newTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertTodo(t, dummyStore.newTodo, newTodo)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("Post invalid content", func(t *testing.T) {
		invalidNewTodo := map[string]string{
			"contnt": "something",
		}
		srv := server.New(dummyStore)

		body := newRequestBody(t, invalidNewTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("Post empty content", func(t *testing.T) {
		emptyNewTodo := types.NewTodo{Content: ""}
		srv := server.New(dummyStore)

		body := newRequestBody(t, emptyNewTodo)
		request, err := newPostTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
}

func TestPut(t *testing.T) {
	t.Run("Update happy path", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 2, Content: "updated content"}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertExist(t, dummyTodos, updatedTodo)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("negative invalid id but valid content", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: -1, Content: "something"}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertNotExist(t, dummyTodos, updatedTodo)
		assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("typo invalid id but valid content", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		typoInvalidUpdatedTodo := struct {
			Ids     int
			Content string
		}{
			2,
			"something",
		}

		body := newRequestBody(t, typoInvalidUpdatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("valid id but empty invalid kcontent", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 2, Content: ""}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertNotExist(t, dummyTodos, updatedTodo)
		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("valid id but invalid typo content", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		typoInvalidContentUpdatedTodo := struct {
			Id     int
			Contnt string
		}{
			2,
			"something",
		}

		body := newRequestBody(t, typoInvalidContentUpdatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("invalid id and invalid content", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: -1, Content: ""}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertNotExist(t, dummyTodos, updatedTodo)
		assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("update id is not exist at current db", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 3, Content: "legit content"}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertNotExist(t, dummyTodos, updatedTodo)
		assertErrMsg(t, response.Body.String(), db.UpdatedIdNotExistErr.Error())
		assertStatus(t, response.Code, http.StatusBadRequest)
	})
	t.Run("update same content and success", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 2, Content: "bar"}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertExist(t, dummyTodos, updatedTodo)
		assertStatus(t, response.Code, http.StatusOK)
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

func newPutTodoRequest(body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest("PUT", "/update", body)
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
		t.Fatalf("Want %v, but got %v", want, got)
	}
}

func assertErrMsg(t testing.TB, got, want string) {
	t.Helper()

	// NOTE: small workaounrd since http.Error() write w with Fprintln()
	comparedGot := strings.TrimSuffix(got, "\n")

	if want != comparedGot {
		t.Fatalf("Want %v, but got %v", want, comparedGot)
	}
}

func assertNoErr(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("didn't expect an error but got one, %v", err)
	}
}

func assertExist(t testing.TB, todos types.Todos, updatedTodo types.UpdateTodo) {
	t.Helper()

	for _, todo := range todos {
		if todo.Id == updatedTodo.Id && todo.Content == updatedTodo.Content {
			return
		}
	}

	t.Fatalf("%v doesn't exist at: %v", updatedTodo, todos)
}

func assertNotExist(t testing.TB, todos types.Todos, updatdedTodo types.UpdateTodo) {
	t.Helper()

	for _, got := range todos {
		if got.Id == updatdedTodo.Id && got.Content == updatdedTodo.Content {
			t.Fatalf("does not expexct %v exist at %v", updatdedTodo, todos)
		}
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("Want %d but got %d", want, got)
	}
}
