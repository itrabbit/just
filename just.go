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
	Version      = "v0.0.6"
	DebugEnvName = "JUST_DEBUG_MODE"
)

// Режим отладки
var debugMutex sync.RWMutex = sync.RWMutex{}
var debugMode bool = true

type IApplication interface {
	IRouter

	Profiler() IProfiler
	Translator() ITranslator
	SerializerManager() ISerializerManager
	TemplatingManager() ITemplatingManager

	SetProfiler(p IProfiler) IApplication
	SetTranslator(t ITranslator) IApplication
	SetNoRouteHandler(handler HandlerFunc) IApplication
	SetNoImplementedHandler(handler HandlerFunc) IApplication

	ServeHTTP(w http.ResponseWriter, req *http.Request)

	Run(address string) error
	RunTLS(address, certFile, keyFile string) error
}

// application - класс приложения
type application struct {
	Router
	pool sync.Pool

	// Стандартные обработчики ошибок
	noRouteHandler       HandlerFunc
	noImplementedHandler HandlerFunc

	// Менеджер сериализаторов с поддержкой многопоточности
	serializerManager serializerManager

	// Менеджер для работы с шаблонами
	templatingManager templatingManager

	// Транлятор локализации i18n
	translator ITranslator

	// Менеджер профилирования
	profiler IProfiler
}

// IApplication::_printWelcomeMessage - выводим приветственное сообщение
func (app *application) _printWelcomeMessage(address string, tls bool) {
	fmt.Print("[WELCOME] Just Web Framework " + Version)
	if tls {
		fmt.Println(" [RUN ON " + address + " / TLS]")
	} else {
		fmt.Println(" [RUN ON " + address + "]")
	}
}

// IApplication::Translator - полчаем транслятор i18n
func (app *application) Translator() ITranslator {
	return app.translator
}

// IApplication::Translator - полчаем транслятор i18n
func (app *application) Profiler() IProfiler {
	return app.profiler
}

// IApplication::Renderer - полчаем отрисовщик по имени
func (app *application) TemplatingManager() ITemplatingManager {
	return &app.templatingManager
}

// IApplication::SetProfiler - назначаем менеджера профилирования
func (app *application) SetProfiler(p IProfiler) IApplication {
	app.profiler = p
	return app
}

// IApplication::SetTranslator - назначаем транслятор i18n
func (app *application) SetTranslator(t ITranslator) IApplication {
	app.translator = t
	return app
}

// IApplication::ServeHTTP - HTTP Handler
func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Усли поступил несуществующий запрос - выходим
	if req == nil || w == nil {
		return
	}
	// Берем контекст из пула
	c := app.pool.Get().(*Context).reset()
	c.Request, c.IsFrozenRequestBody = req, true
	// Передаем контекст в обработчик запросов для заполнения ответа
	app.handleHttpRequest(w, c)
	// Складываем контекст в пул
	app.pool.Put(c)
}

// IApplication::checkMethodForHaveBody - проверка метода с наличием тела запроса по стандарту
func (app *application) checkMethodForHaveBody(method string) bool {
	return method == "POST" || method == "PATCH" || method == "PUT"
}

// IApplication::handleHttpRequest - обрабатываем HTTP запрос используя контекст
func (app *application) handleHttpRequest(w http.ResponseWriter, c *Context) {
	if !c.IsValid() {
		if app.profiler != nil {
			app.profiler.Warning(errors.New("Invalid context"))
		}
		return
	}
	httpMethod, path := c.Request.Method, c.Request.URL.Path
	if app.checkMethodForHaveBody(strings.ToUpper(httpMethod)) && c.Request.Body != nil {
		// TODO: Временное преобразование, исправить в будущем
		// Преобразовываем данные
		if b, _ := ioutil.ReadAll(c.Request.Body); len(b) > 0 {
			c.Request.Body.Close()
			// Новое тело запроса с возможностью сбрасывания позиции чтения
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(b))
		}
	}
	if app.profiler != nil {
		// Фиксация начала обработки запроса
		app.profiler.StartRequest(c.Request)
		// Профилирование выходных данных
		w = &profiledResponseWriter{
			profiler: app.profiler,
			writer:   w,
		}
	}
	// Выполняем handlers из роутеров
	response, existRoute := app.handleRouter(&app.Router, httpMethod, path, c)
	// Если ответ пустой
	if response == nil {
		// Если ответа так и нет, но был найден роут -> выдаем ошибку пустого ответа
		if existRoute {
			// 501 ошибка
			response = app.noImplementedHandler(c)
		} else {
			// Если ничего так и нет, выводим 404 ошибку
			response = app.noRouteHandler(c)
		}
	}
	if app.profiler != nil {
		// Фиксация выбора роута
		app.profiler.SelectRoute(c.Request, c.routeInfo)
	}
	// Отправляем response клиенту
	if response != nil {
		if streamFunc, ok := response.GetStreamHandler(); ok {
			streamFunc(w, c.Request)
		} else {
			if headers := response.GetHeaders(); len(headers) > 0 {
				// Обработка заголовков
				for key, value := range headers {
					if key == "_StrongRedirect" {
						http.Redirect(w, c.Request, value, response.GetStatus())
						return
					}
					if key == "_FilePath" {
						http.ServeFile(w, c.Request, value)
						return
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

// IApplication::handleRouter - обрабатываем HTTP запрос в нужном роуте используя контекст
func (app *application) handleRouter(router *Router, httpMethod, path string, c *Context) (IResponse, bool) {
	if router != nil {
		// Поиск роута
		if router.routes != nil && len(router.routes) > 0 {
			if routes, ok := router.routes[httpMethod]; ok && len(routes) > 0 {
				for _, route := range routes {
					if params, ok := route.CheckPath(path); ok {
						return c.resetRoute(route, params).nextHandler()
					}
				}
			}
		}
		// Поиск следующего роутера
		if router.groups != nil && len(router.groups) > 0 {
			for relativePath, r := range router.groups {
				if strings.Index(path, joinPaths(router.basePath, relativePath)) >= 0 {
					return app.handleRouter(r, httpMethod, path, c)
				}
			}
		}
	}
	return nil, false
}

// IApplication::Run - запуск сервера приложения
func (app *application) Run(address string) error {
	app._printWelcomeMessage(address, false)
	return http.ListenAndServe(address, app)
}

// IApplication::RunTLS - запуск TLS сервера приложения
func (app *application) RunTLS(address, certFile, keyFile string) error {
	app._printWelcomeMessage(address, true)
	return http.ListenAndServeTLS(address, certFile, keyFile, app)
}

// IApplication::SerializerManager - менеджер зериализаторов
func (app *application) SerializerManager() ISerializerManager {
	return &app.serializerManager
}

// IApplication::SetNoRouteHandler - установить обработчик отсутствия роута - 404
func (app *application) SetNoRouteHandler(handler HandlerFunc) IApplication {
	app.noRouteHandler = handler
	return app
}

// IApplication::SetNoImplementedHandler - установить обработчик отсутствия реализации ответа от роута - 501
func (app *application) SetNoImplementedHandler(handler HandlerFunc) IApplication {
	app.noImplementedHandler = handler
	return app
}

// noRouteDefHandler - обработчик ошибки отсутствия роута
func noRouteDefHandler(c *Context) IResponse {
	return c.Serializer().Response(404,
		NewError("404", c.Trans("Route not found")).SetMetadata(H{
			"method": c.Request.Method,
			"path":   c.Request.RequestURI,
		}))
}

// noRouteDefHandler - обработчик ошибки отсутствия реализации
func noImplementedDefHandler(c *Context) IResponse {
	meta := H{
		"method": c.Request.Method,
		"path":   c.Request.RequestURI,
	}
	if c.routeInfo != nil {
		meta["route"] = c.routeInfo.BasePath()
	}
	return c.Serializer().Response(501,
		NewError("501", c.Trans("Response not implemented for current Route")).SetMetadata(meta))
}

func (app *application) initPool() *application {
	app.pool.New = func() interface{} {
		return &Context{app: app}
	}
	return app
}

func (app *application) initSerializers() *application {
	app.serializerManager.SetSerializer("json", []string{
		"application/json",
	}, &JsonSerializer{Charset: "utf-8"}).SetSerializer("xml", []string{
		"text/xml",
		"application/xml",
	}, &XmlSerializer{Charset: "utf-8"}).SetSerializer("form", []string{
		"multipart/form-data",
		"application/x-www-form-urlencoded",
	}, &FormSerializer{Charset: "utf-8"}).SetDefaultName("json")
	return app
}

func (app *application) initRenderers() *application {
	app.TemplatingManager().SetRenderer("html", &HTMLRenderer{Charset: "utf-8"})
	return app
}

// New - создаем приложение
func New() IApplication {
	app := &application{
		Router: Router{
			basePath:        "/",
			handlers:        nil,
			routeParamNames: nil,
			groups:          nil,
			routes:          nil,
		},
		noRouteHandler:       noRouteDefHandler,
		noImplementedHandler: noImplementedDefHandler,
		translator:           &baseTranslator{defaultLocale: "en"},
	}
	return app.initPool().initSerializers().initRenderers()
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
