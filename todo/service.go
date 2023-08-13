package todo

import (
	"github.com/kas2000/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
	"unicode/utf8"
)

type Service interface {
	TodoService
}

type service struct {
	todoRepo TodoRepository
	log      logger.Logger
}

func NewService(todoRepo TodoRepository, log logger.Logger) Service {
	return &service{todoRepo: todoRepo, log: log}
}

func (service *service) CreateTodo(createTodo *CreateTodoDTO) (*Todo, error) {
	if utf8.RuneCountInString(createTodo.Title) > 200 {
		return nil, ErrTitleLengthLimitExceeded
	}
	activeAt, err := time.Parse("2006-01-02", createTodo.ActiveAt)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	return service.todoRepo.Create(&Todo{
		Title:     createTodo.Title,
		Status:    StatusActive,
		ActiveAt:  activeAt,
		CreatedAt: time.Time{},
		UpdatedAt: nil,
	})
}

func (service *service) FindTodo(id primitive.ObjectID) (*Todo, error) {
	return service.todoRepo.FindByID(id)
}

func (service *service) FindTodos(pointers TodoPointers) ([]*Todo, error) {
	var todos []*Todo
	var err error

	switch *pointers.Status {
	case StatusDone:
		todos, err = service.todoRepo.FindAll(pointers)
		if err != nil {
			return nil, err
		}
	case StatusActive:
		comparisonOperator := ComparisonOperatorLTE
		today := time.Now().UTC()
		todos, err = service.todoRepo.FindAll(TodoPointers{
			Status: pointers.Status,
			ActiveAt: &ActiveAtPointers{
				ComparisonOperator: &comparisonOperator,
				ActiveAt:           &today,
			},
		})
		if err != nil {
			return nil, err
		}
	}
	for _, todo := range todos {
		switch todo.ActiveAt.Weekday() {
		case time.Saturday, time.Sunday:
			todo.Title = "ВЫХОДНОЙ - " + todo.Title
		}
	}

	return todos, err
}

func (service *service) UpdateTodo(upd UpdateTodoDTO) error {
	if utf8.RuneCountInString(upd.Title) > 200 {
		return ErrTitleLengthLimitExceeded
	}
	activeAt, err := time.Parse("2006-01-02", upd.ActiveAt)
	if err != nil {
		return ErrInvalidDateFormat
	}
	return service.todoRepo.Update(TodoPointers{
		ID:       &upd.ID,
		Title:    &upd.Title,
		ActiveAt: &ActiveAtPointers{ActiveAt: &activeAt},
	})
}

func (service *service) UpdateTodoStatus(upd TodoPointers) error {
	return service.todoRepo.Update(upd)
}

func (service *service) DeleteTodo(id primitive.ObjectID) error {
	return service.todoRepo.Delete(id)
}
