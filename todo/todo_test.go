package todo

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	command "github.com/kas2000/commandlib"
	"github.com/kas2000/logger"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	validate := validator.New()
	log, _ := logger.New("debug")

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database("regionTaxiDB")
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}
	todoRepo, err := NewTodoRepo(collectionsNames, mongoDB)
	if err != nil {
		log.Fatal("couldn't initialize maintenance repository: " + err.Error())
	}
	service := NewService(todoRepo, log)
	todoCh := command.NewCommandHandler(service)
	todoHttp := NewTodoHttp(log, todoCh, validate, "todo-service")

	reqURL := "/api/todo-list/tasks"

	testCases := []struct {
		title              string
		body               string
		expectedHTTPStatus int
	}{
		{
			title:              "Проверка на корректное создание",
			body:               `{"title":"Купить книгу в декабре","activeAt":"2023-12-04"}`,
			expectedHTTPStatus: 204,
		},
		{
			title:              "Проверка на заполненные поля",
			body:               `{"title":"","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 400,
		},
		{
			title:              "Проверка на дубликаты",
			body:               `{"title":"Купить книгу","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 404,
		},
		{
			title:              "Проверка на длину заголовка",
			body:               `{"title":"Lorem ipsum dolor sit amet consectetur adipisicing elit. Maxime mollitia, molestiae quas vel sint commodi repudiandae consequuntur voluptatum laborum numquam blanditiis harum quisquam eius sed odit fugiat iusto fuga praesentium optio, eaque rerum! Provident similique accusantium nemo autem. Veritatis obcaecati tenetur iure eius earum ut molestias architecto voluptate aliquam nihil, eveniet aliquid culpa officia aut! Impedit sit sunt quaerat, odit, tenetur error, harum nesciunt ipsum debitis quas aliquid. Reprehenderit, quia. Quo neque error repudiandae fuga? Ipsa laudantium molestias eos sapiente officiis modi at sunt excepturi expedita sint? Sed quibusdam recusandae alias error harum maxime adipisci amet laborum. Perspiciatis minima nesciunt dolorem! Officiis iure rerum voluptates a cumque velit quibusdam sed amet tempora. Sit laborum ab, eius fugit doloribus tenetur fugiat, temporibus enim commodi iusto libero magni deleniti quod quam consequuntur! Commodi minima excepturi repudiandae velit hic maxime doloremque. Quaerat provident commodi consectetur veniam similique ad earum omnis ipsum saepe, voluptas, hic voluptates pariatur est explicabo fugiat, dolorum eligendi quam cupiditate excepturi mollitia maiores labore suscipit quas? Nulla, placeat. Voluptatem quaerat non architecto ab laudantium modi minima sunt esse temporibus sint culpa, recusandae aliquam numquam totam ratione voluptas quod exercitationem fuga. Possimus quis earum veniam quasi aliquam eligendi, placeat qui corporis!","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 404,
		},
		{
			title:              "Проверка на валидность даты",
			body:               `{"title":"Купить книгу","activeAt":"2023-13-04"}`,
			expectedHTTPStatus: 404,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(tc.body))
			require.NoError(t, err)

			retData := todoHttp.CreateTodo()(resp, req)
			require.Equal(t, tc.expectedHTTPStatus, retData.StatusCode())
			//Можно было бы проверить response body, но в ТЗ в пункте "В ответе" нет информации что возвращать в теле, указан только статус код
		})
	}
}

func TestDelete(t *testing.T) {
	log, _ := logger.New("debug")
	validate := validator.New()

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database("regionTaxiDB")
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}
	todoRepo, err := NewTodoRepo(collectionsNames, mongoDB)
	if err != nil {
		log.Fatal("couldn't initialize maintenance repository: " + err.Error())
	}
	service := NewService(todoRepo, log)
	todoCh := command.NewCommandHandler(service)
	todoHttp := NewTodoHttp(log, todoCh, validate, "todo-service")

	testCases := []struct {
		title              string
		id                 string
		expectedHTTPStatus int
	}{
		{
			title:              "Проверка на корректное удаление",
			id:                 "64d9fac7fe4ed029b0daf9d0",
			expectedHTTPStatus: 204,
		},
		{
			title:              "Проверка на удаление несуществующей запиcи",
			id:                 "64d9fac7fe4ed029b0daf9d1",
			expectedHTTPStatus: 404,
		},
	}
	reqURL := "/api/todo-list/tasks/"
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
			require.NoError(t, err)

			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			retData := todoHttp.DeleteTodo("id")(resp, req)
			require.Equal(t, tc.expectedHTTPStatus, retData.StatusCode())
		})
	}
}

func TestUpdate(t *testing.T) {
	log, _ := logger.New("debug")
	validate := validator.New()

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database("regionTaxiDB")
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}
	todoRepo, err := NewTodoRepo(collectionsNames, mongoDB)
	if err != nil {
		log.Fatal("couldn't initialize maintenance repository: " + err.Error())
	}
	service := NewService(todoRepo, log)
	todoCh := command.NewCommandHandler(service)
	todoHttp := NewTodoHttp(log, todoCh, validate, "todo-service")

	testCases := []struct {
		title              string
		id                 string
		body               string
		expectedHTTPStatus int
		expectedTitle      string
		expectedActiveAt   string
	}{
		{
			title:              "Проверка на корректное обновление",
			id:                 "64da1f106083a1acd4d8f116",
			body:               `{"title":"Купить книгу - Высоконагруженные приложения","activeAt":"2023-08-05"}`,
			expectedHTTPStatus: 204,
			expectedTitle:      "Купить книгу - Высоконагруженные приложения",
			expectedActiveAt:   "2023-08-05",
		},
		{
			title:              "Проверка на обновление несуществующей запиcи",
			id:                 "64da1f106083a1acd4d8f117",
			body:               `{"title":"Купить книгу","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 404,
		},
		{
			title:              "Проверка на обновление с незаполненными полями",
			id:                 "64d9fac7fe4ed029b0daf9d0",
			body:               `{"title":"","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 400,
		},
		{
			title:              "Проверка на длину заголовка",
			id:                 "64d9fac7fe4ed029b0daf9d0",
			body:               `{"title":"Lorem ipsum dolor sit amet consectetur adipisicing elit. Maxime mollitia, molestiae quas vel sint commodi repudiandae consequuntur voluptatum laborum numquam blanditiis harum quisquam eius sed odit fugiat iusto fuga praesentium optio, eaque rerum! Provident similique accusantium nemo autem. Veritatis obcaecati tenetur iure eius earum ut molestias architecto voluptate aliquam nihil, eveniet aliquid culpa officia aut! Impedit sit sunt quaerat, odit, tenetur error, harum nesciunt ipsum debitis quas aliquid. Reprehenderit, quia. Quo neque error repudiandae fuga? Ipsa laudantium molestias eos sapiente officiis modi at sunt excepturi expedita sint? Sed quibusdam recusandae alias error harum maxime adipisci amet laborum. Perspiciatis minima nesciunt dolorem! Officiis iure rerum voluptates a cumque velit quibusdam sed amet tempora. Sit laborum ab, eius fugit doloribus tenetur fugiat, temporibus enim commodi iusto libero magni deleniti quod quam consequuntur! Commodi minima excepturi repudiandae velit hic maxime doloremque. Quaerat provident commodi consectetur veniam similique ad earum omnis ipsum saepe, voluptas, hic voluptates pariatur est explicabo fugiat, dolorum eligendi quam cupiditate excepturi mollitia maiores labore suscipit quas? Nulla, placeat. Voluptatem quaerat non architecto ab laudantium modi minima sunt esse temporibus sint culpa, recusandae aliquam numquam totam ratione voluptas quod exercitationem fuga. Possimus quis earum veniam quasi aliquam eligendi, placeat qui corporis!","activeAt":"2023-08-04"}`,
			expectedHTTPStatus: 400,
		},
		{
			title:              "Проверка на валидность даты",
			id:                 "64d9fac7fe4ed029b0daf9d0",
			body:               `{"title":"Купить книгу","activeAt":"2023-13-04"}`,
			expectedHTTPStatus: 400,
		},
		{
			title:              "Проверка на дубликаты",
			id:                 "64da1fabd21e112c5bb1c299",
			body:               `{"title":"Купить книгу - Высоконагруженные приложения","activeAt":"2023-08-05"}`,
			expectedHTTPStatus: 400,
		},
	}

	reqURL := "/api/todo-list/tasks/"
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPut, reqURL, strings.NewReader(tc.body))
			require.NoError(t, err)

			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			retData := todoHttp.UpdateTodo("id")(resp, req)
			require.Equal(t, tc.expectedHTTPStatus, retData.StatusCode())

			if tc.title == "Проверка на корректное обновление" {
				id, _ := primitive.ObjectIDFromHex("64da1f106083a1acd4d8f116")
				result, err := service.FindTodo(id)
				if err != nil {
					log.Fatal(err.Error())
				}
				require.Equal(t, result.Title, tc.expectedTitle)
				require.Equal(t, result.ActiveAt, tc.expectedActiveAt)
			}
		})
	}
}

func TestSettingStatusDone(t *testing.T) {
	log, _ := logger.New("debug")
	validate := validator.New()

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database("regionTaxiDB")
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}
	todoRepo, err := NewTodoRepo(collectionsNames, mongoDB)
	if err != nil {
		log.Fatal("couldn't initialize maintenance repository: " + err.Error())
	}
	service := NewService(todoRepo, log)
	todoCh := command.NewCommandHandler(service)
	todoHttp := NewTodoHttp(log, todoCh, validate, "todo-service")

	testCases := []struct {
		title              string
		id                 string
		expectedHTTPStatus int
	}{
		{
			title:              "Проверка на корректное обновление статуса",
			id:                 "64da1f106083a1acd4d8f116",
			expectedHTTPStatus: 204,
		},
		{
			title:              "Проверка на обновление статуса несуществующей запиcи",
			id:                 "64da1f106083a1acd4d8f117",
			expectedHTTPStatus: 404,
		},
	}

	reqURL := "/api/todo-list/tasks/"
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPut, reqURL+tc.id+"/done", nil)
			require.NoError(t, err)
			req = mux.SetURLVars(req, map[string]string{"id": tc.id})

			retData := todoHttp.SetTodoStatusDone("id")(resp, req)
			require.Equal(t, tc.expectedHTTPStatus, retData.StatusCode())

			if tc.title == "Проверка на корректное обновление статуса" {
				id, _ := primitive.ObjectIDFromHex("64da1f106083a1acd4d8f116")
				result, err := todoRepo.FindByID(id)
				if err != nil {
					log.Fatal(err.Error())
				}
				require.Equal(t, result.Status, StatusDone)
			}
		})
	}
}

func TestFind(t *testing.T) {
	log, _ := logger.New("debug")
	validate := validator.New()

	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("couldn't connect to mongodb: " + err.Error())
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Fatal(err.Error())
		}
	}()
	mongoDB := mongoClient.Database("regionTaxiDB")
	collNames, _ := mongoDB.ListCollectionNames(context.TODO(), bson.M{})
	collectionsNames := make(map[string]int)
	for _, collName := range collNames {
		collectionsNames[collName]++
	}
	todoRepo, err := NewTodoRepo(collectionsNames, mongoDB)
	if err != nil {
		log.Fatal("couldn't initialize maintenance repository: " + err.Error())
	}
	service := NewService(todoRepo, log)
	todoCh := command.NewCommandHandler(service)
	todoHttp := NewTodoHttp(log, todoCh, validate, "todo-service")

	testCases := []struct {
		title              string
		queryParam         string
		expectedHTTPStatus int
		expectedResult     GetTodoDTO
	}{
		{
			title:              "Проверка на поиск по status=active",
			queryParam:         "?status=active",
			expectedHTTPStatus: 200,
			expectedResult: GetTodoDTO{
				Title:    "Купить книгу",
				ActiveAt: "2023-08-04",
			},
		},
		{
			title:              "Проверка на поиск по status=done",
			queryParam:         "?status=done",
			expectedHTTPStatus: 200,
			expectedResult: GetTodoDTO{
				Title:    "ВЫХОДНОЙ - Купить книгу - Высоконагруженные приложения",
				ActiveAt: "2023-08-05",
			},
		},
		{
			title:              "Проверка на поиск если status не был указан",
			queryParam:         "",
			expectedHTTPStatus: 200,
			expectedResult: GetTodoDTO{
				Title:    "Купить книгу",
				ActiveAt: "2023-08-04",
			},
		},
	}

	reqURL := "/api/todo-list/tasks"
	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, reqURL+tc.queryParam, nil)
			require.NoError(t, err)

			retData := todoHttp.FindTodos()(resp, req)
			require.Equal(t, tc.expectedHTTPStatus, retData.StatusCode())
			v := retData.Response().([]*GetTodoDTO)
			require.Equal(t, tc.expectedResult.Title, v[0].Title)
			require.Equal(t, tc.expectedResult.ActiveAt, v[0].ActiveAt)
		})
	}
}
