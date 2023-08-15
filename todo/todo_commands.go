package todo

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateTodoCommand struct {
	*CreateTodoDTO
}

func (cmd *CreateTodoCommand) Execute(svc interface{}) (interface{}, error) {
	return svc.(Service).CreateTodo(cmd.CreateTodoDTO)
}

type FindTodoCommand struct {
	ID primitive.ObjectID
}

func (cmd *FindTodoCommand) Execute(svc interface{}) (interface{}, error) {
	return svc.(Service).FindTodo(cmd.ID)
}

type FindTodosCommand struct {
	TodoPointers
}

func (cmd *FindTodosCommand) Execute(svc interface{}) (interface{}, error) {
	return svc.(Service).FindTodos(cmd.TodoPointers)
}

type DeleteTodoCommand struct {
	ID primitive.ObjectID
}

func (cmd *DeleteTodoCommand) Execute(svc interface{}) (interface{}, error) {
	err := svc.(Service).DeleteTodo(cmd.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type UpdateTodoStatusCommand struct {
	TodoPointers
}

func (cmd *UpdateTodoStatusCommand) Execute(svc interface{}) (interface{}, error) {
	err := svc.(Service).UpdateTodoStatus(cmd.TodoPointers)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type UpdateTodoCommand struct {
	UpdateTodoDTO
}

func (cmd *UpdateTodoCommand) Execute(svc interface{}) (interface{}, error) {
	err := svc.(Service).UpdateTodo(cmd.UpdateTodoDTO)
	if err != nil {
		return nil, err
	}
	return nil, nil
}