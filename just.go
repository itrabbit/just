package just

import (
	"fmt"
	"net/http"
	"sync"
)

const (
	Version = "v0.0.1"
)

type (
	// Application - класс приложения
	Application struct {
		pool sync.Pool
	}
)

// Application::ServeHTTP - HTTP Handler
func (app *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Усли поступил несуществующий запрос - выходим
	if req == nil {
		return
	}
	// Берем контекст из пула
	context := app.pool.Get().(*Context)
	// Сбрасываем контекст
	context.reset(w, req)
	// Передаем контекст в обработчик запросов
	app.handleHttpRequest(context)
	// Складываем контекст в пул
	app.pool.Put(context)
}

// Application::handleHttpRequest - обрабатываем HTTP запрос используя контекст
func (app *Application) handleHttpRequest(context *Context) {
	fmt.Println("handleHttpRequest")
}

// Application::Run - запуск сервера приложения
func (app *Application) Run(address string) {
	http.ListenAndServe(address, app)
}

// New - создаем приложение
func New() *Application {
	app := &Application{}
	app.pool.New = func() interface{} {
		return &Context{app: app}
	}
	return app
}
