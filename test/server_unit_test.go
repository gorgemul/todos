package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorgemul/todos/pkg/db"
	"github.com/gorgemul/todos/pkg/server"
	"github.com/gorgemul/todos/types"
)

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
	t.Run("happy path", func(t *testing.T) {
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
	t.Run("happy path", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 2, Content: "updated content"}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertIdAndContentExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
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

		assertIdAndContentNotExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
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
	t.Run("valid id but empty invalid content", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		updatedTodo := types.UpdateTodo{Id: 2, Content: ""}

		body := newRequestBody(t, updatedTodo)
		request, err := newPutTodoRequest(body)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertIdAndContentNotExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
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

		assertIdAndContentNotExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
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

		assertIdAndContentNotExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
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

		assertIdAndContentExist(t, dummyTodos, updatedTodo.Id, updatedTodo.Content)
		assertStatus(t, response.Code, http.StatusOK)
	})
}

func TestDelete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		deleteId := 2

		request, err := newDeleteTodoRequest(deleteId)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertIdNotExist(t, dummyTodos, deleteId)
	})
	t.Run("negative invalid delete id", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		deleteId := -5

		request, err := newDeleteTodoRequest(deleteId)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
		assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
	})
	t.Run("delete id valid but not exist at current db", func(t *testing.T) {
		store := &stubStore{Todos: dummyTodos}
		srv := server.New(store)
		deleteId := 3

		request, err := newDeleteTodoRequest(deleteId)
		assertNoErr(t, err)
		response := httptest.NewRecorder()

		srv.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
		assertErrMsg(t, response.Body.String(), db.DeleteIdNotExistErr.Error())
	})
}
