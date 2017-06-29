package just

import (
	"regexp"
	"strings"
)

var (
	rxPathReplaceParams = regexp.MustCompile(`(\{[^/]*?\})`)
	rxPathFindParams    = regexp.MustCompile(`\{([^/]*?)\}`)
)

type (
	// Req - текстовые рекомендации для роутинга
	Req map[string]string

	// IRouter - интерфейс роутера
	IRouter interface {
		IRoute
		Group(string, ...HandlerFunc) IRouter
	}
	// IRoute - интерфейс роута
	IRoute interface {
		IRouteInfo

		// Использовать middleware
		Use(...HandlerFunc) IRoute

		// Обработка запросов к приложению сервера
		Handle(string, string, Req, ...HandlerFunc) IRoute
		Any(string, Req, ...HandlerFunc) IRoute
		GET(string, Req, ...HandlerFunc) IRoute
		POST(string, Req, ...HandlerFunc) IRoute
		DELETE(string, Req, ...HandlerFunc) IRoute
		PATCH(string, Req, ...HandlerFunc) IRoute
		PUT(string, Req, ...HandlerFunc) IRoute
		OPTIONS(string, Req, ...HandlerFunc) IRoute
		HEAD(string, Req, ...HandlerFunc) IRoute

		CheckPath(string) (map[string]string, bool)
	}
	IRouteInfo interface {
		BasePath() string
		Handlers() []HandlerFunc
		HandlerByIndex(index int) HandlerFunc
	}
	// Router - роутер
	Router struct {
		basePath        string              // Путь который обрабатываем роут
		rxPath          *regexp.Regexp      // Регулярное выражение для проверки пути
		handlers        []HandlerFunc       // Список обработчиков, в том числе и middleware
		routeParamNames []string            // Список обнаруженных параметров в пути роута
		groups          map[string]*Router  // Роутеры (map[relativePath]*Router)
		routes          map[string][]IRoute // Роуты с группировкой по методу (map[httpMethod][]IRoute)
	}
)

func (r *Router) handle(httpMethod string, relativePath string, req Req, handlers []HandlerFunc) IRoute {
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
						routeParamNames[i] = strings.TrimSpace(param[1])
						if pattern, ok := req[strings.TrimSpace(param[1])]; ok {
							regExpPattern = strings.Replace(regExpPattern, param[0], "("+pattern+")", 1)
							continue
						}
					} else {
						routeParamNames[i] = strings.TrimSpace(param[0])
					}
					regExpPattern = strings.Replace(regExpPattern, param[0], "([^/\\\\]*?)", 1)
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
		handlers:        handlers,
		routeParamNames: routeParamNames,
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
		handlers:        handlers,
		routeParamNames: nil,
		groups:          nil,
		routes:          nil,
	}
	if r.groups == nil {
		r.groups = make(map[string]*Router)
	}
	r.groups[relativePath] = group
	return group
}

func (r *Router) Handle(httpMethod, relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("HTTP method [" + httpMethod + "] not valid")
	}
	return r.handle(httpMethod, relativePath, req, handlers)
}

func (r *Router) POST(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("POST", relativePath, req, handlers)
}

func (r *Router) GET(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("GET", relativePath, req, handlers)
}

func (r *Router) DELETE(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("DELETE", relativePath, req, handlers)
}

func (r *Router) PATCH(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("PATCH", relativePath, req, handlers)
}

func (r *Router) PUT(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("PUT", relativePath, req, handlers)
}

func (r *Router) OPTIONS(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("OPTIONS", relativePath, req, handlers)
}

func (r *Router) HEAD(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	return r.handle("HEAD", relativePath, req, handlers)
}

func (r *Router) Any(relativePath string, req Req, handlers ...HandlerFunc) IRoute {
	r.handle("GET", relativePath, req, handlers)
	r.handle("POST", relativePath, req, handlers)
	r.handle("PUT", relativePath, req, handlers)
	r.handle("PATCH", relativePath, req, handlers)
	r.handle("DELETE", relativePath, req, handlers)
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

func (r *Router) Handlers() []HandlerFunc {
	return r.handlers
}

func (r *Router) HandlerByIndex(index int) HandlerFunc {
	if index >= 0 && len(r.handlers) > index {
		return r.handlers[index]
	}
	return nil
}
