package todo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Status    string             `json:"status" bson:"status"`
	ActiveAt  time.Time          `json:"active_at" bson:"active_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at" bson:"updated_at"`
}

type TodoPointers struct {
	ID        *primitive.ObjectID
	Title     *string
	Status    *string
	ActiveAt  *ActiveAtPointers
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type CreateTodoDTO struct {
	Title    string `json:"title" validate:"required"`
	ActiveAt string `json:"active_at" validate:"required"`
}

type UpdateTodoDTO struct {
	ID       primitive.ObjectID `json:"id"`
	Title    string             `json:"title" validate:"required"`
	ActiveAt string             `json:"active_at" validate:"required"`
}

type ActiveAtPointers struct {
	ComparisonOperator *string
	ActiveAt           *time.Time
}

type TodoRepository interface {
	Create(todo *Todo) (*Todo, error)
	FindByID(id primitive.ObjectID) (*Todo, error)
	FindAll(pointers TodoPointers) ([]*Todo, error)
	Update(upd TodoPointers) error
	Delete(id primitive.ObjectID) error
}

type TodoService interface {
	CreateTodo(todo *CreateTodoDTO) (*Todo, error)
	FindTodo(id primitive.ObjectID) (*Todo, error)
	FindTodos(pointers TodoPointers) ([]*Todo, error)
	UpdateTodo(upd UpdateTodoDTO) error
	UpdateTodoStatus(upd TodoPointers) error
	DeleteTodo(id primitive.ObjectID) error
}

const (
	StatusActive = "ACTIVE"
	StatusDone   = "DONE"

	ComparisonOperatorEQ  = "EQ"
	ComparisonOperatorGT  = "GT"
	ComparisonOperatorGTE = "GTE"
	ComparisonOperatorLT  = "LT"
	ComparisonOperatorLTE = "LTE"
)

var (
	ErrTodoNotFound              = errors.New("todo not found.")
	ErrUnknownComparisonOperator = errors.New("unknown comparison operator.")
	ErrNothingToUpdate           = errors.New("nothing to update.")
	ErrTitleLengthLimitExceeded  = errors.New("title length limit exceeded.")
	ErrInvalidDateFormat         = errors.New("invalid date format.")
)
