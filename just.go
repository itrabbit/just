package just

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

const (
	Version      = "v0.1.16"
	DebugEnvName = "JUST_DEBUG_MODE"
)

// Debug mode.
var (
	debugMutex = sync.RWMutex{}
	debugMode  = true
)

// JUST Application interface.
type IApplication interface {
	IRouter

	Profiler() IProfiler
	Translator() ITranslator
	SerializerManager() ISerializerManager
	TemplatingManager() ITemplatingManager

	LocalDo(req *http.Request) IResponse

	SetProfiler(p IProfiler) IApplication
	SetTranslator(t ITranslator) IApplication
	SetNoRouteHandler(handler HandlerFunc) IApplication
	SetNoImplementedHandler(handler HandlerFunc) IApplication

	ServeHTTP(w http.ResponseWriter, req *http.Request)

	Run(address string) error
	RunTLS(address, certFile, keyFile string) error
}

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

func (app *application) printWelcomeMessage(address string, tls bool) {
	fmt.Print("[WELCOME] Just Web Framework ", Version)
	if tls {
		fmt.Println(" [RUN ON", address, "/ TLS]")
	} else {
		fmt.Println(" [RUN ON", address+"]")
	}
}

func (app *application) Translator() ITranslator {
	return app.translator
}

func (app *application) Profiler() IProfiler {
	return app.profiler
}

func (app *application) TemplatingManager() ITemplatingManager {
	return &app.templatingManager
}

func (app *application) SetProfiler(p IProfiler) IApplication {
	app.profiler = p
	return app
}

func (app *application) SetTranslator(t ITranslator) IApplication {
	app.translator = t
	return app
}

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

func (app *application) checkMethodForHaveBody(method string) bool {
	return method == "POST" || method == "PATCH" || method == "PUT"
}

func (app *application) handleHttpRequest(w http.ResponseWriter, c *Context) {
	if !c.IsValid() {
		if app.profiler != nil {
			app.profiler.Warning(errors.New("Invalid context"))
		}
		return
	}
	httpMethod, path := c.Request.Method, c.Request.URL.Path
	if app.checkMethodForHaveBody(httpMethod) && c.Request.Body != nil {
		// Преобразовываем данные
		if b, _ := ioutil.ReadAll(c.Request.Body); len(b) > 0 {
			c.Request.Body.Close()
			// Новое тело запроса с возможностью сбрасывания позиции чтения
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(b))
		}
	}
	if app.profiler != nil {
		// Фиксация начала обработки запроса
		app.profiler.OnStartRequest(c.Request)
		// Профилирование выходных данных
		w = &profiledResponseWriter{
			profiler: app.profiler,
			writer:   w,
		}
	}
	// Recover
	defer func() {
		if rvr := recover(); rvr != nil {

			fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
			debug.PrintStack()

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			if app.profiler != nil {
				app.profiler.Error(errors.New("invalid response, recover panic"))
			}
		}
	}()
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
			// но перед этим прогоняем все обработчики
			if response = c.resetRoute(app, nil).Next(); response == nil {
				response = app.noRouteHandler(c)
			}
		}
	}
	if app.profiler != nil {
		// Фиксация выбора роута
		app.profiler.OnSelectRoute(c.Request, c.routeInfo)
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
	// Если ничего не смогли сделать, выдаем 405 ошибку
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	if app.profiler != nil {
		app.profiler.Error(http.ErrNotSupported)
	}
}

// Локальный вызов запроса внутри движка
func (app *application) LocalDo(req *http.Request) IResponse {
	if req == nil {
		return nil
	}
	c := app.pool.Get().(*Context).reset()
	defer app.pool.Put(c)

	c.Request, c.IsFrozenRequestBody = req, true
	c.isLocalRequest = true

	httpMethod, path := c.Request.Method, c.Request.URL.Path
	if app.checkMethodForHaveBody(httpMethod) && c.Request.Body != nil {
		if b, _ := ioutil.ReadAll(c.Request.Body); len(b) > 0 {
			c.Request.Body.Close()
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(b))
		}
	}
	defer func() {
		if rvr := recover(); rvr != nil {
			fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
			debug.PrintStack()
		}
	}()
	response, _ := app.handleRouter(&app.Router, httpMethod, path, c)
	return response
}

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
			for _, r := range router.groups {
				if strings.Index(r.basePath, "{") >= 0 && r.rxPath != nil {
					if _, ok := r.CheckPath(path); ok {
						return app.handleRouter(r, httpMethod, path, c)
					}
					continue
				}
				if strings.Index(path, r.basePath) >= 0 {
					return app.handleRouter(r, httpMethod, path, c)
				}
			}
		}
	}
	return nil, false
}

func (app *application) Run(address string) error {
	app.printWelcomeMessage(address, false)
	return http.ListenAndServe(address, app)
}

func (app *application) RunTLS(address, certFile, keyFile string) error {
	app.printWelcomeMessage(address, true)
	return http.ListenAndServeTLS(address, certFile, keyFile, app)
}

func (app *application) SerializerManager() ISerializerManager {
	return &app.serializerManager
}

func (app *application) SetNoRouteHandler(handler HandlerFunc) IApplication {
	app.noRouteHandler = handler
	return app
}

func (app *application) SetNoImplementedHandler(handler HandlerFunc) IApplication {
	app.noImplementedHandler = handler
	return app
}

func noRouteDefHandler(c *Context) IResponse {
	return c.Serializer().Response(404,
		NewError("404", c.Trans("Route not found")).SetMetadata(H{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}))
}

func noImplementedDefHandler(c *Context) IResponse {
	meta := H{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
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

// Set default serializer to just application (json, xml, form-data, x-www-form-urlencoded)
func SetDefSerializers(app IApplication) IApplication {
	app.SerializerManager().SetSerializer("json", []string{
		"application/json",
	}, &JsonSerializer{Ch: "utf-8"}).SetSerializer("xml", []string{
		"text/xml",
		"application/xml",
	}, &XmlSerializer{Ch: "utf-8"}).SetSerializer("form", []string{
		"multipart/form-data",
		"application/x-www-form-urlencoded",
	}, &FormSerializer{Ch: "utf-8", OnlyDeserialize: true}).SetDefaultName("json")
	return app
}

// Set default renderers (HTML UTF-8)
func SetDefRenderers(app IApplication) IApplication {
	app.TemplatingManager().SetRenderer("html", &HTMLRenderer{Charset: "utf-8"})
	return app
}

// Create new clear JUST application.
// Without serializers and renderers.
func NewClear() IApplication {
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
	return app.initPool()
}

// Create new default JUST application.
func New() IApplication {
	return SetDefSerializers(SetDefRenderers(NewClear()))
}

// Change debug mode
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
