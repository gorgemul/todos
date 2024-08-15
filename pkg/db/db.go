package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gorgemul/todos/types"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var (
	noContext            = context.Background()
	UpdatedIdNotExistErr = errors.New("Updated todo id is not exist!")
	DeleteIdNotExistErr  = errors.New("Deleted todo id is not exist!")
)

type DBStore struct {
	*pgx.Conn
}

func (db *DBStore) GetTodos() (types.Todos, error) {
	rows, err := db.Query(noContext, "SELECT * FROM todozz")

	if err != nil {
		return nil, err
	}

	var todos types.Todos

	for rows.Next() {
		var todo types.Todo
		if err := rows.Scan(&todo.Id, &todo.Content, &todo.CreatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (db *DBStore) PostTodo(content string) error {
	_, err := db.Exec(noContext, "INSERT INTO todozz (content) VALUES ($1);", content)

	if err != nil {
		return err
	}

	return nil
}

func (db *DBStore) UpdateTodo(id int, content string) error {
	result, err := db.Exec(noContext, "UPDATE todozz SET content = $1 WHERE id = $2", content, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return UpdatedIdNotExistErr
	}

	return nil
}

func (db *DBStore) DeleteTodo(id int) error {
	result, err := db.Exec(noContext, "DELETE FROM todozz WHERE id = $1", id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return DeleteIdNotExistErr
	}

	return nil
}

func New() (*DBStore, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(noContext, os.Getenv("TODO_DB"))
	if err != nil {
		return nil, fmt.Errorf("problem connecting to db: %v", err)
	}

	return &DBStore{conn}, nil
}
