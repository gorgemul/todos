package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/gorgemul/todos/pkg/db"
	"github.com/gorgemul/todos/pkg/server"
	"github.com/gorgemul/todos/types"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	database *pgxpool.Pool
)

type idCounter struct {
	current int
}

func TestMain(m *testing.M) {
	pool, err := createDockerPool()
	if err != nil {
		log.Fatal(err)
	}

	resource, err := createDBContainer(pool)
	if err != nil {
		log.Fatal(err)
	}

	databaseUrl := getHostDBUrl(resource)

	if err := tryConnectDBUntil(120*time.Second, resource, pool, databaseUrl); err != nil {
		log.Fatal(err)
	}

	defer database.Close()
	defer removeContainerFromPool(pool, resource)

	m.Run()
}

func TestHappyPath(t *testing.T) {
	m, err := createTables()
	assertNoErr(t, err)

	defer dropAllTables(m)

	dbStore := &db.DBStore{Pool: database}
	srv := server.New(dbStore)
	expected := types.Todos{}
	counter := &idCounter{current: 1}

	t.Run("init state", func(t *testing.T) {
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("add three entries", func(t *testing.T) {
		expected = add(t, srv, counter, "new 1", expected)
		expected = add(t, srv, counter, "new 2", expected)
		expected = add(t, srv, counter, "new 3", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("update one entry", func(t *testing.T) {
		expected = updateById(t, srv, 2, "legit content", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("Delete one entry", func(t *testing.T) {
		expected = deleteById(t, srv, 2, expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("After delete add several entries to make sure id is searial incremantal form", func(t *testing.T) {
		expected = add(t, srv, counter, "new 4", expected)
		expected = add(t, srv, counter, "new 5", expected)
		expected = add(t, srv, counter, "new 6", expected)
		expected = add(t, srv, counter, "new 7", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})
}

func TestErrorPath(t *testing.T) {
	m, err := createTables()
	assertNoErr(t, err)

	defer dropAllTables(m)

	dbStore := &db.DBStore{Pool: database}
	srv := server.New(dbStore)
	expected := types.Todos{}
	counter := &idCounter{current: 1}

	t.Run("add empty content to db", func(t *testing.T) {
		expected = add(t, srv, counter, "", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("add mistype content to db", func(t *testing.T) {
		addMistypeContent(t, srv, "something")
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("update negative id and invalid id", func(t *testing.T) {
		expected = updateById(t, srv, -1, "something", expected)
		expected = updateById(t, srv, 1, "something", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("add one entry to db", func(t *testing.T) {
		expected = add(t, srv, counter, "legit content", expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("update mistype id from db", func(t *testing.T) {
		updateMistypeId(t, srv, 1, "something")
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("update mistype content from db", func(t *testing.T) {
		updateMistypeContent(t, srv, 1, "something")
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("update mistype id and content from db", func(t *testing.T) {
		updateMistypeIdAndContent(t, srv, 1, "something")
		got := get(t, srv)
		assertTodos(t, got, expected)
	})

	t.Run("delete negative id and invalid id", func(t *testing.T) {
		expected = deleteById(t, srv, -1, expected)
		expected = deleteById(t, srv, 2, expected)
		got := get(t, srv)
		assertTodos(t, got, expected)
	})
}

func createDockerPool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func createDBContainer(pool *dockertest.Pool) (*dockertest.Resource, error) {
	// pulls an image, create a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=admin",
			"POSTGRES_DB=todos",
			"listen_address = '5438'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func getHostDBUrl(resource *dockertest.Resource) string {
	hostPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://admin:secret@%s/todos?sslmode=disable", hostPort)

	return databaseUrl
}

func tryConnectDBUntil(t time.Duration, resource *dockertest.Resource, pool *dockertest.Pool, databaseUrl string) error {
	log.Println("Connecting to database on url", databaseUrl)

	seconds := uint(t.Seconds())
	resource.Expire(seconds)
	pool.MaxWait = t

	if err := pool.Retry(func() error {
		var err error
		database, err = pgxpool.New(context.Background(), databaseUrl)
		if err != nil {
			return err
		}
		return database.Ping(context.Background())
	}); err != nil {
		return err
	}

	return nil
}

func removeContainerFromPool(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func createTables() (*migrate.Migrate, error) {
	connInfo := database.Config().ConnString()
	m, err := migrate.New(
		"file://../internal/db/migrations",
		connInfo,
	)
	if err != nil {
		return nil, err
	}
	if err := m.Up(); err != nil {
		return nil, err
	}

	return m, nil
}

func dropAllTables(m *migrate.Migrate) {
	m.Down()
}

func get(t *testing.T, srv *server.Server) types.Todos {
	getRequest, err := newGetTodoRequest()
	assertNoErr(t, err)

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, getRequest)
	assertStatus(t, response.Code, http.StatusOK)

	return getTodosFromResponse(t, response)
}

func add(t *testing.T, srv *server.Server, counter *idCounter, content string, expected types.Todos) types.Todos {
	newTodo := types.NewTodo{Content: content}
	requestBody := newRequestBody(t, newTodo)
	postRequest, err := newPostTodoRequest(requestBody)
	assertNoErr(t, err)

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, postRequest)
	if response.Code != http.StatusOK {
		assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
		assertStatus(t, response.Code, http.StatusBadRequest)
		return expected
	}
	newTodos := append(expected, types.Todo{Id: counter.current, Content: content, CreatedAt: dummyTime})
	counter.current++
	return newTodos
}

func addMistypeContent(t *testing.T, srv *server.Server, content string) {
	invalidNewTodo := map[string]string{
		"cont": content,
	}
	requestBody := newRequestBody(t, invalidNewTodo)
	postRequest, err := newPostTodoRequest(requestBody)
	assertNoErr(t, err)

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, postRequest)

	assertStatus(t, response.Code, http.StatusBadRequest)
	assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
}

func updateById(t *testing.T, srv *server.Server, id int, content string, expected types.Todos) types.Todos {
	updatedTodo := types.UpdateTodo{Id: id, Content: content}
	body := newRequestBody(t, updatedTodo)
	putRequest, err := newPutTodoRequest(body)
	assertNoErr(t, err)

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, putRequest)
	if response.Code != http.StatusOK {
		assertStatus(t, response.Code, http.StatusBadRequest)
		assertErrMsgIn(t, response.Body.String(), server.InvalidIdErrMsg, db.UpdatedIdNotExistErr.Error())
	}
	for i, v := range expected {
		if v.Id == id {
			expected[i].Content = content
			break
		}
	}
	return expected
}

func updateMistypeId(t *testing.T, srv *server.Server, id int, content string) {
	invalidUpdatedTodo := struct {
		Ids     int
		Content string
	}{
		id,
		content,
	}
	requestBody := newRequestBody(t, invalidUpdatedTodo)
	putRequest, err := newPutTodoRequest(requestBody)
	assertNoErr(t, err)
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, putRequest)
	assertStatus(t, response.Code, http.StatusBadRequest)
	assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
}

func updateMistypeContent(t *testing.T, srv *server.Server, id int, content string) {
	invalidUpdatedTodo := struct {
		Id     int
		contnt string
	}{
		id,
		content,
	}
	requestBody := newRequestBody(t, invalidUpdatedTodo)
	putRequest, err := newPutTodoRequest(requestBody)
	assertNoErr(t, err)
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, putRequest)
	assertStatus(t, response.Code, http.StatusBadRequest)
	assertErrMsg(t, response.Body.String(), server.InvalidContentErrMsg)
}

func updateMistypeIdAndContent(t *testing.T, srv *server.Server, id int, content string) {
	invalidUpdatedTodo := struct {
		Ids    int
		contnt string
	}{
		id,
		content,
	}
	requestBody := newRequestBody(t, invalidUpdatedTodo)
	putRequest, err := newPutTodoRequest(requestBody)
	assertNoErr(t, err)
	response := httptest.NewRecorder()
	srv.ServeHTTP(response, putRequest)
	assertStatus(t, response.Code, http.StatusBadRequest)
	assertErrMsg(t, response.Body.String(), server.InvalidIdErrMsg)
}

func deleteById(t *testing.T, srv *server.Server, id int, expected types.Todos) types.Todos {
	response := httptest.NewRecorder()
	request, err := newDeleteTodoRequest(id)
	assertNoErr(t, err)

	srv.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		assertStatus(t, response.Code, http.StatusBadRequest)
		assertErrMsgIn(t, response.Body.String(), server.InvalidIdErrMsg, db.DeleteIdNotExistErr.Error())
	}
	expected = slices.DeleteFunc(expected, func(todo types.Todo) bool {
		return todo.Id == id
	})
	return expected
}
