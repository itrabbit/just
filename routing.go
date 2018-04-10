package just

import (
	"fmt"
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
	patternParamFileExt = "(\\.[A-Za-z0-9]+|)"
	patternParamBoolean = "(1|0|t|f|true|false|T|F|TRUE|FALSE)"
	patternParamUUID    = "([a-fA-F0-9]{8}-?[a-f0-9]{4}-?[1-5][a-fA-F0-9]{3}-?[89abAB][a-fA-F0-9]{3}-?[a-fA-F0-9]{12})"
	patternParamRID     = "([0-9a-z]{20})"
	patternHex          = "((0[xX])?[0-9a-fA-F]+)"
)

// Method of processing a request or middleware.
type HandlerFunc func(*Context) IResponse

// Interface information on route.
type IRouteInfo interface {
	BasePath() string
	CountHandlers() int
	HandlerByIndex(index int) (HandlerFunc, bool)
}

// Route interface.
type IRoute interface {
	IRouteInfo

	// Use middleware.
	Use(...HandlerFunc) IRoute

	// Processing of requests to the application server.
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

// Router interface.
type IRouter interface {
	IRoute
	Group(string, ...HandlerFunc) IRouter
}

// Base Router struct.
type Router struct {
	basePath        string              // The way that process route.
	rxPath          *regexp.Regexp      // The regular expression used to validate the path.
	handlers        []HandlerFunc       // The list of processors, including middleware available for this route.
	routeParamNames []string            // A list of detected parameters in their path.
	parent          *Router             // A pointer to the parent router.
	groups          map[string]*Router  // Routers (map[relativePath]*Router).
	routes          map[string][]IRoute // Routes with grouping by method (map[httpMethod][]IRoute).
}

func connectHandlersByRouter(r *Router, handlers []HandlerFunc) []HandlerFunc {
	if r != nil && r.handlers != nil {
		return append(append(make([]HandlerFunc, 0, 0), r.handlers...), handlers...)
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
								case "p", "path":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamPath, 1)
								case "hex":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternHex, 1)
								case "rid":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamRID, 1)
								case "uuid":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamUUID, 1)
								case "i", "int", "integer":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamInteger, 1)
								case "f", "number", "float":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamFloat, 1)
								case "b", "bool", "boolean":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamBoolean, 1)
								case "f.e", "file.ext":
									regExpPattern = strings.Replace(regExpPattern, param[0], patternParamFileExt, 1)
								default:
									{
										findPattern = false
										if begin, end := strings.Index(req, "("), strings.LastIndex(req, ")"); begin > 0 && end > begin {
											if t := strings.TrimSpace(req[:begin]); len(t) > 0 {
												if findPattern = t == "regexp" || t == "enum"; findPattern && len(strings.TrimSpace(req[begin+1:end])) > 0 {
													switch t {
													case "rgx", "regexp":
														regExpPattern = strings.Replace(regExpPattern, param[0], strings.TrimSpace(req[begin:]), 1)
													case "e", "enum":
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
			if IsDebug() {
				fmt.Println("[DEBUG] Registration", httpMethod, "route regexp:", regExpPattern, routeParamNames)
			}
			rxPath, err = regexp.Compile("^" + regExpPattern + "$")
			if err != nil {
				panic(err)
			}
		}
	} else if IsDebug() {
		fmt.Println("[DEBUG] Registration", httpMethod, "route:", basePath)
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

// Use middleware.
func (r *Router) Use(middleware ...HandlerFunc) IRoute {
	if r.handlers == nil {
		r.handlers = make([]HandlerFunc, 0)
	}
	r.handlers = append(r.handlers, middleware...)
	return r
}

// Create group router.
// The group does not support regular expressions, text only.
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

// Create a HTTP request handler.
func (r *Router) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoute {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("HTTP method [" + httpMethod + "] not valid")
	}
	return r.handle(httpMethod, relativePath, handlers)
}

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead of the Router's NotFound handler.
// To use the operating system's file system implementation, use:
// `router.Static("/static", "/var/www")`
func (r *Router) Static(relativePath, root string) IRoute {
	return r.StaticFS(relativePath, http.Dir(root))
}

// StaticFile registers a single route in order to server a single file of the local filesystem.
// `router.StaticFile("favicon.ico", "./resources/favicon.ico")`
func (r *Router) StaticFile(relativePath, filePath string) IRoute {
	handler := func(c *Context) IResponse {
		return FileResponse(filePath)
	}
	return r.GET(relativePath, handler).HEAD(relativePath, handler)
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
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

// POST is a shortcut for router.Handle("POST", path, handlers...).
func (r *Router) POST(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodPost, relativePath, handlers)
}

// GET is a shortcut for router.Handle("GET", path, handlers...).
func (r *Router) GET(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodGet, relativePath, handlers)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers...).
func (r *Router) DELETE(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodDelete, relativePath, handlers)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handlers...).
func (r *Router) PATCH(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodPatch, relativePath, handlers)
}

// PUT is a shortcut for router.Handle("PUT", path, handlers...).
func (r *Router) PUT(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodPut, relativePath, handlers)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers...).
func (r *Router) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodOptions, relativePath, handlers)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handlers...).
func (r *Router) HEAD(relativePath string, handlers ...HandlerFunc) IRoute {
	return r.handle(http.MethodHead, relativePath, handlers)
}

// Any registers a route that matches all the HTTP methods. GET, POST, PUT, PATCH, DELETE.
func (r *Router) ANY(relativePath string, handlers ...HandlerFunc) IRoute {
	r.handle(http.MethodGet, relativePath, handlers)
	r.handle(http.MethodPost, relativePath, handlers)
	r.handle(http.MethodPut, relativePath, handlers)
	r.handle(http.MethodPatch, relativePath, handlers)
	r.handle(http.MethodDelete, relativePath, handlers)
	return r
}

func (r *Router) CheckPath(path string) (map[string]string, bool) {
	if r.rxPath != nil {
		if r.rxPath.MatchString(path) {
			if indexes := r.rxPath.FindStringSubmatchIndex(path); len(indexes) > 2 && len(indexes)%2 == 0 {
				values := make([]string, 0)
				for e, i := 0, 2; i < len(indexes); i += 2 {
					if indexes[i] >= e {
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
