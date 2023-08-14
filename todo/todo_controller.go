package todo

import "github.com/kas2000/http"

type todoController struct {
	server *http.Server
	http   *TodoHttp
	prefix string
}

func NewTodoController(server *http.Server, http *TodoHttp, prefix string) *todoController {
	return &todoController{
		server: server,
		http:   http,
		prefix: prefix,
	}
}

func (tc *todoController) Bind() {
	srvr := *tc.server
	srvr.Handle("POST", tc.prefix+"/todo-list/tasks", tc.http.CreateTodo())
	srvr.Handle("GET", tc.prefix+" /todo-list/tasks", tc.http.FindTodos())
	srvr.Handle("GET", tc.prefix+"/todo-list/tasks/{id}", tc.http.FindTodo("id"))
	srvr.Handle("PUT", tc.prefix+"/todo-list/tasks/{id}", tc.http.UpdateTodo("id"))
	srvr.Handle("PUT", tc.prefix+"/todo-list/tasks/{id}/done", tc.http.SetTodoStatusDone("id"))
	srvr.Handle("DELETE", tc.prefix+"/todo-list/tasks/{id}", tc.http.DeleteTodo("id"))
}
