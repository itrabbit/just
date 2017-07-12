package just

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	Version      = "v0.0.2"
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

func (app *Application) defResponse404(context *Context) IResponse {
	return context.ResponseDataFast(404,
		NewError("404", "Route not found").SetMetadata(H{
			"method": context.Request.Method,
			"path":   context.Request.RequestURI,
		}))
}

func (app *Application) defResponse501(context *Context) IResponse {
	meta := H{
		"method": context.Request.Method,
		"path":   context.Request.RequestURI,
	}
	if context.routeInfo != nil {
		meta["route"] = context.routeInfo.BasePath()
	}
	return context.ResponseDataFast(501,
		NewError("501", "Response not implemented for current Route").SetMetadata(meta))
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
	response, existRoute := app.handleRouter(&app.Router, httpMethod, path, context)
	// Если ответ пустой
	if response == nil {
		// Если ответа так и нет, но был найден роут -> выдаем ошибку пустого ответа
		if existRoute {
			// 501 ошибка
			response = app.defResponse501(context)
		} else {
			// Если ничего так и нет, выводим 404 ошибку
			response = app.defResponse404(context)
		}
	}
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
	panic(errors.New("Empty Response"))
}

// Application::handleRouter - обрабатываем HTTP запрос в нужном роуте используя контекст
func (app *Application) handleRouter(router *Router, httpMethod, path string, context *Context) (IResponse, bool) {
	if router != nil {
		// Работа с событиями
		if router.handlers != nil && len(router.handlers) > 0 {
			if resp, ok := context.resetRoute(router, nil).Next(); ok && resp != nil {
				return resp, ok
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
	return nil, false
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
