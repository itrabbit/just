package just

import (
	"net/http"
	"path"
	"regexp"
	"strings"
	"unicode"
)

var (
	rxPathFindParams = regexp.MustCompile(`\{([^/]*?)\}`)
)

const (
	patternParamPath    = "(.*?)"
	patternParamString  = "([^/\\\\]*?)"
	patternParamFloat   = "([+-]?(\\d*[.])?\\d+)"
	patternParamInteger = "([+-]?\\d+)"
	patternParamBoolean = "(1|0|t|f|true|false|T|F|TRUE|FALSE)"
	patternParamUUID    = "([a-fA-F0-9]{8}-?[a-f0-9]{4}-?[1-5][a-fA-F0-9]{3}-?[89abAB][a-fA-F0-9]{3}-?[a-fA-F0-9]{12})"
)

// HandlerFunc - метод обработки запроса или middleware
type HandlerFunc func(*Context) IResponse

type IRouteInfo interface {
	BasePath() string
	CountHandlers() int
	HandlerByIndex(index int) (HandlerFunc, bool)
}

// IRoute - интерфейс роута
type IRoute interface {
	IRouteInfo

	// Использовать middleware
	Use(...HandlerFunc) IRoute

	// Обработка запросов к приложению сервера
	Handle(string, string, ...HandlerFunc) IRoute
	ANY(string, ...HandlerFunc) IRoute
	GET(string, ...HandlerFunc) IRoute
	POST(string, ...HandlerFunc) IRoute
	DELETE(string, ...HandlerFunc) IRoute
	PATCH(string, ...HandlerFunc) IRoute
	PUT(string, ...HandlerFunc) IRoute
	OPTIONS(string, ...HandlerFunc) IRoute
	HEAD(string, ...HandlerFunc) IRoute

	StaticFile(string, string) IRoute

	Static(string, string) IRoute
	StaticFS(string, http.FileSystem) IRoute

	CheckPath(string) (map[string]string, bool)
}

// IRouter - интерфейс роутера
type IRouter interface {
	IRoute
	Group(string, ...HandlerFunc) IRouter
}

// Router - роутер
type Router struct {
	basePath        string              // Путь который обрабатываем роут
	rxPath          *regexp.Regexp      // Регулярное выражение для проверки пути
	handlers        []HandlerFunc       // Список обработчиков, в том числе и middleware доступные в данном роуте
	routeParamNames []string            // Список обнаруженных параметров в пути роута
	parent          *Router             // Указатель на родительский роутер
	groups          map[string]*Router  // Роутеры (map[relativePath]*Router)
	routes          map[string][]IRoute // Роуты с группировкой по методу (map[httpMethod][]IRoute)
}

func connectHandlersByRouter(r *Router, handlers []HandlerFunc) []HandlerFunc {
	if r != nil && r.handlers != nil {
		return append(r.handlers, handlers...)
	}
	return handlers
}

func (r *Router) handle(httpMethod string, relativePath string, handlers []HandlerFunc) IRoute {
	if r.routes == nil {
		r.routes = make(map[string][]IRoute)
	}
	if _, ok := r.routes[httpMethod]; !ok {
		r.routes[httpMethod] = make([]IRoute, 0)
	}
	// Формируем условие для совпадения пути
	basePath := joinPaths(r.basePath, relativePath)

	var routeParamNames []string = nil
	var rxPath *regexp.Regexp = nil

	if strings.Index(basePath, "{") >= 0 {
		params := rxPathFindParams.FindAllStringSubmatch(basePath, -1)
		if len(params) > 0 {
			// Формируем полные рекомендации и по ним строим пути
			routeParamNames = make([]string, len(params))
			regExpPattern := basePath
			for i, param := range params {
				if len(param) > 0 {
					if len(param) > 1 {
						// Анализ параметра
						if pos := strings.Index(param[1], ":"); pos > 0 {
							routeParamNames[i] = strings.TrimSpace(param[1][0:pos])
							if req := strings.TrimSpace(param[1][pos+1:]); len(req) > 1 {
								// Анализ рекомендаций параметра
								findPattern := true
								switch req {
								case "path":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamPath, 1)
								case "uuid":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamUUID, 1)
								case "integer":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamInteger, 1)
								case "float":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamFloat, 1)
								case "boolean":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamBoolean, 1)
								default:
									{
										findPattern = false
										if begin, end := strings.Index(req, "("), strings.LastIndex(req, ")"); begin > 0 && end > begin {
											if t := strings.ToLower(strings.TrimSpace(req[:begin])); len(t) > 0 {
												if findPattern = t == "regexp" || t == "enum"; findPattern && len(strings.TrimSpace(req[begin+1:end])) > 0 {
													switch t {
													case "regexp":
														regExpPattern = strings.Replace(regExpPattern, param[0], strings.TrimSpace(req[begin:]), 1)
													case "enum":
														regExpPattern = strings.Replace(regExpPattern, param[0], "("+strings.Join(strings.FieldsFunc(req[begin+1:end], func(c rune) bool {
															return !unicode.IsLetter(c) && !unicode.IsNumber(c)
														}), "|")+")", 1)
													}
												}
											}
										}
									}
								}
								if findPattern {
									continue
								}
							}
						} else {
							routeParamNames[i] = strings.TrimSpace(param[1])
						}
					} else {
						routeParamNames[i] = strings.TrimSpace(param[0])
					}
					regExpPattern = strings.Replace(regExpPattern, param[0], patternParamString, 1)
				}
			}
			var err error
			rxPath, err = regexp.Compile("^" + regExpPattern + "$")
			if err != nil {
				panic(err)
			}
		}
	}
	r.routes[httpMethod] = append(r.routes[httpMethod], &Router{
		basePath:        basePath,
		rxPath:          rxPath,
		handlers:        connectHandlersByRouter(r, handlers),
		routeParamNames: routeParamNames,
		parent:          r,
		groups:          nil,
		routes:          nil,
	})
	return r
}

func (r *Router) Use(middleware ...HandlerFunc) IRoute {
	if r.handlers == nil {
		r.handlers = make([]HandlerFunc, 0)
	}
	r.handlers = append(r.handlers, middleware...)
	return r
}

func (r *Router) Group(relativePath string, handlers ...HandlerFunc) IRouter {
	group := &Router{
		basePath:        joinPaths(r.basePath, relativePath),
		rxPath:          nil, // TODO: Пока роутер не может иметь параметры в пути
		handlers:        connectHandlersByRouter(r, handlers),
		routeParamNames: nil,
		parent:          r,
		groups:          nil,
		routes:          nil,
	}
	if r.groups == nil {
		r.groups = make(map[string]*Router)
	}
	r.groups[relativePath] = group
	return group
}

func (r *Router) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoute {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("HTTP method [" + httpMethod + "] not valid")
	}
	return r.handle(httpMethod, relativePath, handlers)
}

func (r *Router) Static(relativePath, root string) IRoute {
	return r.StaticFS(relativePath, http.Dir(root))
}

func (r *Router) StaticFile(relativePath, filePath string) IRoute {
	handler := func(c *Context) IResponse {
		return FileResponse(filePath)
	}
	return r.GET(relativePath, handler).HEAD(relativePath, handler)
}

func (r *Router) StaticFS(relativePath string, fs http.FileSystem) IRoute {
	fileServer := http.StripPrefix(joinPaths(r.basePath, relativePath), http.FileServer(fs))
	handler := func(c *Context) IResponse {
		return StreamResponse(func(w http.ResponseWriter, r *http.Request) {
			fileServer.ServeHTTP(w, r)
		})
	}
	urlPattern := path.Join(relativePath, "/{filepath:path}")
	return r.GET(urlPattern).HEAD(urlPattern, handler)
}

func (r *Router) POST(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("POST", relativePath, handlers)
}

func (r *Router) GET(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("GET", relativePath, handlers)
}

func (r *Router) DELETE(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("DELETE", relativePath, handlers)
}

func (r *Router) PATCH(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("PATCH", relativePath, handlers)
}

func (r *Router) PUT(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("PUT", relativePath, handlers)
}

func (r *Router) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("OPTIONS", relativePath, handlers)
}

func (r *Router) HEAD(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle("HEAD", relativePath, handlers)
}

func (r *Router) ANY(relativePath string, handlers ...HandlerFunc) IRoute {
	r.handle("GET", relativePath, handlers)
	r.handle("POST", relativePath, handlers)
	r.handle("PUT", relativePath, handlers)
	r.handle("PATCH", relativePath, handlers)
	r.handle("DELETE", relativePath, handlers)
	return r
}

func (r *Router) CheckPath(path string) (map[string]string, bool) {
	if r.rxPath != nil {
		if r.rxPath.MatchString(path) {
			if indexes := r.rxPath.FindStringSubmatchIndex(path); len(indexes) > 2 && len(indexes)%2 == 0 {
				values := make([]string, 0)
				for e, i := 0, 2; i < len(indexes); i += 2 {
					if indexes[i] > e {
						values = append(values, path[indexes[i]:indexes[i+1]])
						e = indexes[i+1]
					}
				}
				if len(values) >= len(r.routeParamNames) {
					params := make(map[string]string)
					for i := 0; i < len(r.routeParamNames); i++ {
						params[r.routeParamNames[i]] = values[i]
					}
					return params, true
				}
			}
			return nil, false
		}
	}
	return nil, strings.Compare(path, r.basePath) == 0
}

func (r *Router) BasePath() string {
	return r.basePath
}

func (r *Router) CountHandlers() int {
	return len(r.handlers)
}

func (r *Router) HandlerByIndex(index int) (HandlerFunc, bool) {
	if index >= 0 && index < len(r.handlers) {
		return r.handlers[index], true
	}
	return nil, false
}
