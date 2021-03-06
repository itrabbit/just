package just

import (
	"fmt"
	"net/http"
	"testing"
)

func BenchmarkOneRoute(B *testing.B) {
	app := New()
	app.GET("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkOneRoute", B, app, "GET", "/ping")
}

func BenchmarkRecoveryMiddleware(B *testing.B) {
	app := New()
	app.GET("/", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkRecoveryMiddleware", B, app, "GET", "/")
}

func BenchmarkLoggerMiddleware(B *testing.B) {
	app := New()
	app.GET("/", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkLoggerMiddleware", B, app, "GET", "/")
}

func BenchmarkManyHandlers(B *testing.B) {
	app := New()
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkManyHandlers", B, app, "GET", "/ping")
}

func Benchmark5Params(B *testing.B) {
	app := New()
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/param/{param1}/{params2}/{param3}/{param4}/{param5:integer}", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("Benchmark5Params", B, app, "GET", "/param/path/to/parameter/john/12345")
}

func BenchmarkOneRouteJSON(B *testing.B) {
	app := New()
	data := struct {
		Status string `json:"status"`
	}{"ok"}
	app.GET("/json", func(c *Context) IResponse {
		return JsonResponse(200, data)
	})
	runRequest("BenchmarkOneRouteJSON", B, app, "GET", "/json")
}

func BenchmarkOneRouteHTML(B *testing.B) {
	app := New()
	app.TemplatingManager().Renderer("html").LoadTemplateGlob("index", "<html><body><h1>{{.}}</h1></body></html>")
	app.GET("/html", func(c *Context) IResponse {
		return c.Renderer("html").Response(200, "index", "hola")
	})
	runRequest("BenchmarkOneRouteHTML", B, app, "GET", "/html")
}

func BenchmarkOneRouteSet(B *testing.B) {
	app := New()
	app.GET("/ping", func(c *Context) IResponse {
		c.Set("key", "value")
		return &Response{Status: 200}
	})
	runRequest("BenchmarkOneRouteSet", B, app, "GET", "/ping")
}

func BenchmarkLocalDo(B *testing.B) {
	app := New()
	app.GET("/ping", func(c *Context) IResponse {
		c.Set("key", "value")
		return &Response{Status: 200}
	})
	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		panic(err)
	}
	B.ReportAllocs()
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		app.LocalDo(req)
	}
}

func BenchmarkOneRouteString(B *testing.B) {
	app := New()
	app.GET("/text", func(c *Context) IResponse {
		return &Response{Status: 200, Bytes: []byte("this is a plain text")}
	})
	runRequest("BenchmarkOneRouteString", B, app, "GET", "/text")
}

func BenchmarkManyRoutesFist(B *testing.B) {
	app := New()
	app.ANY("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkManyRoutesFist", B, app, "GET", "/ping")
}

func BenchmarkManyRoutesLast(B *testing.B) {
	app := New()
	app.ANY("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("BenchmarkManyRoutesLast", B, app, "DELETE", "/ping")
}

func Benchmark404(B *testing.B) {
	app := New()
	app.ANY("/something", func(c *Context) IResponse { return &Response{Status: 200} })
	app.SetNoRouteHandler(func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("Benchmark404", B, app, "GET", "/ping")
}

func Benchmark404Many(B *testing.B) {
	app := New()
	app.GET("/", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/path/to/something", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/post/{id:integer}", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/view/{id:integer}", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/favicon.ico", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/robots.txt", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/delete/{id:integer}", func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/user/{id:integer}/{mode}", func(c *Context) IResponse { return &Response{Status: 200} })

	app.SetNoRouteHandler(func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest("Benchmark404Many", B, app, "GET", "/viewfake")
}

type mockWriter struct {
	headers http.Header
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		http.Header{},
	}
}

func (m *mockWriter) Header() (h http.Header) {
	return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockWriter) WriteHeader(int) {}

func runRequest(name string, B *testing.B, r IApplication, method, path string) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	w := newMockWriter()
	fmt.Println(name + ":")
	B.ReportAllocs()
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		r.ServeHTTP(w, req)
	}
}

func init() {
	SetDebugMode(false)
}
