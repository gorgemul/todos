package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func New() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("TODO_DB"))
	if err != nil {
		return nil, fmt.Errorf("problem connecting to db: %v", err)
	}

	return conn, nil
}
