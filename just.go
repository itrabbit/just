package just

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const (
	Version = "v0.0.1"
)

// HandlerFunc - метод обработки запроса или middleware
type HandlerFunc func(*Context) IResponse

// Application - класс приложения
type Application struct {
	Router
	pool sync.Pool
}

// Application::ServeHTTP - HTTP Handler
func (app *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Усли поступил несуществующий запрос - выходим
	if req == nil {
		return
	}
	// Берем контекст из пула
	context := app.pool.Get().(*Context)
	// Сбрасываем контекст
	context.reset(req)
	// Передаем контекст в обработчик запросов для заполнения ответа
	app.handleHttpRequest(w, context)
	// Складываем контекст в пул
	app.pool.Put(context)
}

// Application::handleHttpRequest - обрабатываем HTTP запрос используя контекст
func (app *Application) handleHttpRequest(w http.ResponseWriter, context *Context) {
	if !context.IsValid() {
		return
	}
	httpMethod, path := context.Request().Method, context.Request().URL.Path

	// Проходимся по пути и обходим все связанные группы роутинга
	fmt.Println(httpMethod, path)

	// Выполняем handlers из роутеров
	response := app.handleRouter(&app.Router, httpMethod, path, context)

	// Отправляем response клиенту
	if response != nil {
		w.WriteHeader(response.GetStatus())
		if headers := response.GetHeaders(); len(headers) > 0 {
			for key, value := range headers {
				w.Header().Set(key, value)
			}
		}
		w.Write(response.GetData())
		return
	}
	// Если ничего так и нет, выводим 404 ошибку

}

// Application::handleRouter - обрабатываем HTTP запрос в нужном роуте используя контекст
func (app *Application) handleRouter(router *Router, httpMethod, path string, context *Context) IResponse {
	if router != nil {
		// Работа с событиями
		if router.handlers != nil && len(router.handlers) > 0 {
			if resp := context.resetRoute(router, nil).Next(); resp != nil {
				return resp
			}
		}
		// Поиск роута
		if router.routes != nil && len(router.routes) > 0 {
			if routes, ok := router.routes[httpMethod]; ok && len(routes) > 0 {
				for _, route := range routes {
					if params, ok := route.CheckPath(path); ok {
						return context.resetRoute(route, params).Next()
					}
				}
			}
		}
		// Поиск следующего роутера
		if router.groups != nil && len(router.groups) > 0 {
			for relativePath, r := range router.groups {
				if strings.Index(path, joinPaths(router.basePath, relativePath)) >= 0 {
					return app.handleRouter(r, httpMethod, path, context)
				}
			}
		}
	}
	return nil
}

// Application::Run - запуск сервера приложения
func (app *Application) Run(address string) {
	http.ListenAndServe(address, app)
}

// New - создаем приложение
func New() *Application {
	app := &Application{
		Router: Router{
			basePath:        "/",
			handlers:        nil,
			routeParamNames: nil,
			groups:          nil,
			routes:          nil,
		},
	}
	app.pool.New = func() interface{} {
		return &Context{app: app}
	}
	return app
}
