package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorgemul/todos/pkg/db"
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
	lenBeforeDelete := len(s.Todos)

	s.Todos = slices.DeleteFunc(s.Todos, func(todo types.Todo) bool {
		return todo.Id == id
	})

	if len(s.Todos) == lenBeforeDelete {
		return db.DeleteIdNotExistErr
	}

	return nil
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

func newDeleteTodoRequest(deleteId int) (*http.Request, error) {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("/delete/%d", deleteId), nil)
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
		t.Fatalf("Want %v (type %T), but got %v (type %T)", want, want, got, got)
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

func assertIdAndContentExist(t testing.TB, todos types.Todos, id int, content string) {
	t.Helper()

	for _, todo := range todos {
		if todo.Id == id && todo.Content == content {
			return
		}
	}

	t.Fatalf("{Id: %d, Content: %s} doesn't exist at: %v", id, content, todos)
}

func assertIdAndContentNotExist(t testing.TB, todos types.Todos, id int, content string) {
	t.Helper()

	for _, got := range todos {
		if got.Id == id && got.Content == content {
			t.Fatalf("{Id: %d, Content: %s} not expect to exist at %v", id, content, todos)
		}
	}
}

func assertIdNotExist(t testing.TB, todos types.Todos, id int) {
	t.Helper()

	for _, got := range todos {
		if got.Id == id {
			t.Fatalf("id: %d not expect to exist at %v", id, todos)
		}
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("Want %d but got %d", want, got)
	}
}

func assertTodos(t testing.TB, gots, wants types.Todos) {
	t.Helper()

	for i, got := range gots {
		want := wants[i]
		if got.Content != want.Content || got.Id != want.Id {
			t.Fatalf("want: %v, but got: %v", wants, gots)
		}
	}
}

func assertErrMsgIn(t testing.TB, got string, wants ...string) {
	t.Helper()

	// NOTE: small workaounrd since http.Error() write w with Fprintln()
	comparedGot := strings.TrimSuffix(got, "\n")

	for _, want := range wants {
		if comparedGot == want {
			return
		}
	}

	t.Fatalf("got: %s, not inside wants: %v", comparedGot, wants)
}
