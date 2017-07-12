package just

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	Version      = "v0.0.1"
	DebugEnvName = "JUST_DEBUG_MODE"
)

// Режим отладки
var debugMutex sync.RWMutex = sync.RWMutex{}
var debugMode bool = true

// Application - класс приложения
type Application struct {
	Router
	pool sync.Pool

	// Менеджер сериализаторов с поддержкой многопоточности
	serializerManager serializerManager
}

// Application::ServeHTTP - HTTP Handler
func (app *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Усли поступил несуществующий запрос - выходим
	if req == nil || w == nil {
		return
	}
	// Берем контекст из пула
	context := app.pool.Get().(*Context).reset()
	context.Request, context.IsFrozenRequestBody = req, true
	// Передаем контекст в обработчик запросов для заполнения ответа
	app.handleHttpRequest(w, context)
	// Складываем контекст в пул
	app.pool.Put(context)
}

func (app *Application) checkMethodForHaveBody(method string) bool {
	return method == "POST" || method == "PATCH" || method == "PUT"
}

// Application::handleHttpRequest - обрабатываем HTTP запрос используя контекст
func (app *Application) handleHttpRequest(w http.ResponseWriter, context *Context) {
	if !context.IsValid() {
		return
	}
	httpMethod, path := context.Request.Method, context.Request.URL.Path
	if app.checkMethodForHaveBody(strings.ToUpper(httpMethod)) && context.Request.Body != nil {
		// TODO: Временное преобразование, исправить в будущем
		// Преобразовываем данные
		if b, _ := ioutil.ReadAll(context.Request.Body); len(b) > 0 {
			context.Request.Body.Close()
			// Новое тело запроса с возможностью сбрасывания позиции чтения
			context.Request.Body = ioutil.NopCloser(bytes.NewReader(b))
		}
	}
	// Выполняем handlers из роутеров
	response := app.handleRouter(&app.Router, httpMethod, path, context)

	// Отправляем response клиенту
	if response != nil {
		if headers := response.GetHeaders(); len(headers) > 0 {
			for key, value := range headers {
				if key == "Content-Type" {
					if strings.Index(value, ";") < 0 {
						value = strings.TrimSpace(value) + "; charset=utf-8"
					}
				}
				w.Header().Set(key, value)
			}
		}
		w.WriteHeader(response.GetStatus())
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
func (app *Application) Run(address string) error {
	return http.ListenAndServe(address, app)
}

// Application::GetSerializerManager - менеджер зериализаторов
func (app *Application) GetSerializerManager() ISerializerManager {
	return &app.serializerManager
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
	app.serializerManager.SetSerializer("json", []string{
		"application/json",
	}, &JsonSerializer{}).SetSerializer("xml", []string{
		"text/xml",
		"application/xml",
	}, &XmlSerializer{}).SetSerializer("form", []string{
		"multipart/form-data",
		"application/x-www-form-urlencoded",
	}, &FormSerializer{}).SetNameDefaultSerializer("json")
	return app
}

func SetDebugMode(value bool) {
	debugMutex.RLock()
	defer debugMutex.RUnlock()
	debugMode = value
}

func IsDebug() bool {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	return debugMode
}

func init() {
	if value, err := strconv.ParseBool(os.Getenv(DebugEnvName)); err == nil {
		SetDebugMode(value)
	}
}
