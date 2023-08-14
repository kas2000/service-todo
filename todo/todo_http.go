package todo

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	httpLib "github.com/kas2000/http"
	"github.com/kas2000/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"net/http"
)
import command "github.com/kas2000/commandlib"

type TodoHttp struct {
	log        logger.Logger
	ch         command.CommandHandler
	validate   *validator.Validate
	systemName string
}

func NewTodoHttp(log logger.Logger, ch command.CommandHandler, validate *validator.Validate, systemName string) *TodoHttp {
	return &TodoHttp{
		log:        log,
		ch:         ch,
		validate:   validate,
		systemName: systemName,
	}
}

func (factory *TodoHttp) CreateTodo() httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return httpLib.BadRequest(100, "Error reading request body: "+err.Error(), factory.systemName)
		}

		var todo CreateTodoDTO
		err = json.Unmarshal(body, &todo)
		if err != nil {
			return httpLib.BadRequest(110, "Error unmarshalling: "+err.Error(), factory.systemName)
		}

		err = factory.validate.Struct(todo)
		if err != nil {
			return httpLib.BadRequest(120, err.Error(), factory.systemName)
		}

		cmd := CreateTodoCommand{
			CreateTodoDTO: &todo,
		}

		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			return httpLib.NotFound(130, err.Error(), factory.systemName) //Почему в тз написано возвращаем 404? Разве не 500 должна быть?
			//return httpLib.InternalServer(150, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusNoContent, resp, nil) //Почему в тз написано возвращаем 204?
		//return httpLib.NewResponse(http.StatusCreated, resp, nil) Разве не 201 должна быть?
	}
}

func (factory *TodoHttp) UpdateTodo(idParameter string) httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {
		vars := mux.Vars(r)
		id, found := vars[idParameter]
		if !found {
			return httpLib.BadRequest(140, "no subject id.", factory.systemName)
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return httpLib.BadRequest(150, err.Error(), factory.systemName)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return httpLib.BadRequest(160, "Error reading request body: "+err.Error(), factory.systemName)
		}
		var upd UpdateTodoDTO
		err = json.Unmarshal(body, &upd)
		if err != nil {
			return httpLib.BadRequest(170, "Error unmarshalling: "+err.Error(), factory.systemName)
		}
		upd.ID = objID

		err = factory.validate.Struct(upd)
		if err != nil {
			return httpLib.BadRequest(180, err.Error(), factory.systemName)
		}

		cmd := UpdateTodoCommand{
			UpdateTodoDTO: upd,
		}
		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			switch err {
			case ErrTodoNotFound:
				return httpLib.NotFound(190, err.Error(), factory.systemName)
			}
			return httpLib.InternalServer(200, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusNoContent, resp, nil) //Почему в тз написано возвращаем 204?
	}
}

func (factory *TodoHttp) SetTodoStatusDone(idParameter string) httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {
		vars := mux.Vars(r)
		id, found := vars[idParameter]
		if !found {
			return httpLib.BadRequest(140, "no subject id.", factory.systemName)
		}
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return httpLib.BadRequest(150, err.Error(), factory.systemName)
		}

		status := StatusDone
		cmd := UpdateTodoStatusCommand{
			TodoPointers: TodoPointers{
				ID:     &objID,
				Status: &status,
			},
		}
		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			switch err {
			case ErrTodoNotFound:
				return httpLib.NotFound(190, err.Error(), factory.systemName)
			}
			return httpLib.InternalServer(200, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusNoContent, resp, nil) //Почему в тз написано возвращаем 204?
	}
}

func (factory *TodoHttp) DeleteTodo(idParameter string) httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {
		vars := mux.Vars(r)
		id, found := vars[idParameter]
		if !found {
			return httpLib.BadRequest(210, "no subject id", factory.systemName)
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return httpLib.BadRequest(220, err.Error(), factory.systemName)
		}

		cmd := DeleteTodoCommand{ID: objID}

		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			switch err {
			case ErrTodoNotFound:
				return httpLib.NotFound(230, err.Error(), factory.systemName)
			}
			return httpLib.InternalServer(240, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusNoContent, resp, nil)
	}
}

func (factory *TodoHttp) FindTodo(idParameter string) httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {
		vars := mux.Vars(r)
		id, found := vars[idParameter]
		if !found {
			return httpLib.BadRequest(250, "No subject id", factory.systemName)
		}

		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return httpLib.BadRequest(260, err.Error(), factory.systemName)
		}

		cmd := FindTodoCommand{ID: objID}

		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			switch err {
			case ErrTodoNotFound:
				return httpLib.NotFound(270, err.Error(), factory.systemName)
			}
			return httpLib.InternalServer(280, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusOK, resp, nil)
	}
}

func (factory *TodoHttp) FindTodos() httpLib.Endpoint {
	return func(w http.ResponseWriter, r *http.Request) httpLib.Response {

		var pointers TodoPointers
		if r.URL.Query().Has("status") {
			status := r.URL.Query().Get("status")
			pointers.Status = &status
		} else {
			statusActive := StatusActive
			pointers.Status = &statusActive
		}

		if r.URL.Query().Has("title") {
			title := r.URL.Query().Get("title")
			pointers.Title = &title
		}

		cmd := FindTodosCommand{
			TodoPointers: pointers,
		}

		resp, err := factory.ch.ExecuteCommand(&cmd)
		if err != nil {
			return httpLib.InternalServer(290, err.Error(), factory.systemName)
		}
		return httpLib.NewResponse(http.StatusOK, resp, nil)
	}
}
