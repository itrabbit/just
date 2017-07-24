package just

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	Version      = "v0.0.5"
	DebugEnvName = "JUST_DEBUG_MODE"
)

// Режим отладки
var debugMutex sync.RWMutex = sync.RWMutex{}
var debugMode bool = true

// Application - класс приложения
type Application struct {
	Router
	pool sync.Pool

	// Кодировка ответа по умолчанию
	defCharset string

	// Стандартные обработчики ошибок
	noRoute       HandlerFunc
	noImplemented HandlerFunc

	// Менеджер сериализаторов с поддержкой многопоточности
	serializerManager serializerManager

	// Менеджер для работы с шаблонами
	templatingManager templatingManager

	// Менеджер профилирования
	profiler IProfiler
}

// Application::_printWelcomeMessage - выводим приветственное сообщение
func (app *Application) _printWelcomeMessage(address string, tls bool) {
	fmt.Print("[WELCOME] Just Web Framework " + Version)
	if tls {
		fmt.Println(" [RUN ON " + address + " / TLS]")
	} else {
		fmt.Println(" [RUN ON " + address + "]")
	}
}

// Application::Renderer - полчаем отрисовщик по имени
func (app *Application) TemplatingManager() ITemplatingManager {
	return &app.templatingManager
}

// Application::SetProfiler - назначаем менеджера профилирования
func (app *Application) SetProfiler(p IProfiler) {
	app.profiler = p
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

// Application::checkMethodForHaveBody - проверка метода с наличием тела запроса по стандарту
func (app *Application) checkMethodForHaveBody(method string) bool {
	return method == "POST" || method == "PATCH" || method == "PUT"
}

// Application::handleHttpRequest - обрабатываем HTTP запрос используя контекст
func (app *Application) handleHttpRequest(w http.ResponseWriter, context *Context) {
	if !context.IsValid() {
		if app.profiler != nil {
			app.profiler.Warning(errors.New("Invalid context"))
		}
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
	if app.profiler != nil {
		// Фиксация начала обработки запроса
		app.profiler.StartRequest(context.Request)
		// Профилирование выходных данных
		w = &profiledResponseWriter{
			profiler: app.profiler,
			writer:   w,
		}
	}
	// Выполняем handlers из роутеров
	response, existRoute := app.handleRouter(&app.Router, httpMethod, path, context)
	// Если ответ пустой
	if response == nil {
		// Если ответа так и нет, но был найден роут -> выдаем ошибку пустого ответа
		if existRoute {
			// 501 ошибка
			response = app.noImplemented(context)
		} else {
			// Если ничего так и нет, выводим 404 ошибку
			response = app.noRoute(context)
		}
	}
	if app.profiler != nil {
		// Фиксация выбора роута
		app.profiler.SelectRoute(context.Request, context.routeInfo)
	}
	// Отправляем response клиенту
	if response != nil {
		if streamFunc, ok := response.GetStreamHandler(); ok {
			streamFunc(w, context.Request)
		} else {
			if headers := response.GetHeaders(); len(headers) > 0 {
				for key, value := range headers {
					if key == "_StrongRedirect" {
						http.Redirect(w, context.Request, value, response.GetStatus())
						return
					}
					if key == "_FilePath" {
						http.ServeFile(w, context.Request, value)
						return
					}
					if key == "Content-Type" {
						if strings.Index(value, ";") < 0 {
							value = strings.TrimSpace(value) + "; charset=" + app.defCharset
						}
					}
					w.Header().Set(key, value)
				}
			}
			w.WriteHeader(response.GetStatus())
			w.Write(response.GetData())
		}
		return
	}
	// Если ничего не смогли сделать, выдаем 500 ошибку
	w.WriteHeader(500)
	w.Write([]byte("500 - Internal Server Error.\r\nThe server could not process your request, or the response could not be sent."))
	if app.profiler != nil {
		app.profiler.Error(errors.New("Invalid response"))
	}
}

// Application::handleRouter - обрабатываем HTTP запрос в нужном роуте используя контекст
func (app *Application) handleRouter(router *Router, httpMethod, path string, context *Context) (IResponse, bool) {
	if router != nil {
		// Поиск роута
		if router.routes != nil && len(router.routes) > 0 {
			if routes, ok := router.routes[httpMethod]; ok && len(routes) > 0 {
				for _, route := range routes {
					if params, ok := route.CheckPath(path); ok {
						return context.resetRoute(route, params).nextHandler()
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
	if len(app.defCharset) < 2 {
		app.defCharset = "utf-8"
	}
	app._printWelcomeMessage(address, false)
	return http.ListenAndServe(address, app)
}

// Application::RunTLS - запуск TLS сервера приложения
func (app *Application) RunTLS(address, certFile, keyFile string) error {
	if len(app.defCharset) < 2 {
		app.defCharset = "utf-8"
	}
	app._printWelcomeMessage(address, true)
	return http.ListenAndServeTLS(address, certFile, keyFile, app)
}

// Application::SerializerManager - менеджер зериализаторов
func (app *Application) SerializerManager() ISerializerManager {
	return &app.serializerManager
}

// Application::NoRoute - установить обработчик отсутствия роута - 404
func (app *Application) NoRoute(handler HandlerFunc) *Application {
	app.noRoute = handler
	return app
}

// Application::NoImplemented - установить обработчик отсутствия реализации ответа от роута - 501
func (app *Application) NoImplemented(handler HandlerFunc) *Application {
	app.noImplemented = handler
	return app
}

// noRouteDefHandler - обработчик ошибки отсутствия роута
func noRouteDefHandler(context *Context) IResponse {
	return context.ResponseDataFast(404,
		NewError("404", "Route not found").SetMetadata(H{
			"method": context.Request.Method,
			"path":   context.Request.RequestURI,
		}))
}

// noRouteDefHandler - обработчик ошибки отсутствия реализации
func noImplementedDefHandler(context *Context) IResponse {
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

func (app *Application) InitPool() *Application {
	app.pool.New = func() interface{} {
		return &Context{app: app}
	}
	return app
}

func (app *Application) InitSerializers() *Application {
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

func (app *Application) InitRenderers() *Application {
	app.TemplatingManager().SetRenderer("html", &HTMLRenderer{})
	return app
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

		defCharset: "utf-8",

		noRoute:       noRouteDefHandler,
		noImplemented: noImplementedDefHandler,
	}
	return app.InitPool().InitSerializers().InitRenderers()
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
