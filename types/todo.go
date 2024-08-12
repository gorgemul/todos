package types

import "time"

type Todos []Todo

type Todo struct {
	Id        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type NewTodo struct {
	Content string
}

type UpdateTodo struct {
	Id      int
	Content string
}
