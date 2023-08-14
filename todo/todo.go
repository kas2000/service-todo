package todo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Status    string             `json:"status" bson:"status"`
	ActiveAt  time.Time          `json:"activeAt" bson:"active_at"`
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

type GetTodoDTO struct {
	Title    string `json:"title"`
	ActiveAt string `json:"activeAt"`
}

type CreateTodoDTO struct {
	Title    string `json:"title" validate:"required"`
	ActiveAt string `json:"activeAt" validate:"required"`
}

type UpdateTodoDTO struct {
	ID       primitive.ObjectID `json:"id"`
	Title    string             `json:"title" validate:"required"`
	ActiveAt string             `json:"activeAt" validate:"required"`
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
	CreateTodo(todo *CreateTodoDTO) (*GetTodoDTO, error)
	FindTodo(id primitive.ObjectID) (*GetTodoDTO, error)
	FindTodos(pointers TodoPointers) ([]*GetTodoDTO, error)
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
	ErrTodoAlreadyExists         = errors.New("todo already exists.")
	ErrUnknownComparisonOperator = errors.New("unknown comparison operator.")
	ErrNothingToUpdate           = errors.New("nothing to update.")
	ErrTitleLengthLimitExceeded  = errors.New("title length limit exceeded.")
	ErrInvalidDateFormat         = errors.New("invalid date format.")
)

func ToDateString(date time.Time) string {
	var month string
	var day string
	if date.Month() < 10 {
		month = month + "0"
	}
	month = month + strconv.Itoa(int(date.Month()))
	if date.Day() < 10 {
		day = day + "0"
	}
	day = day + strconv.Itoa(date.Day())
	dateString := strconv.Itoa(date.Year()) + "-" + month + "-" + day
	return dateString
}