package just

import (
	"net/http"
	"testing"
)

func BenchmarkOneRoute(B *testing.B) {
	app := New()
	app.GET("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/ping")
}

func BenchmarkRecoveryMiddleware(B *testing.B) {
	app := New()
	app.GET("/", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/")
}

func BenchmarkLoggerMiddleware(B *testing.B) {
	app := New()
	app.GET("/", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/")
}

func BenchmarkManyHandlers(B *testing.B) {
	app := New()
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/ping")
}

func Benchmark5Params(B *testing.B) {
	app := New()
	app.Use(func(c *Context) IResponse { return &Response{Status: 200} })
	app.GET("/param/{param1}/{params2}/{param3}/{param4}/{param5}", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/param/path/to/parameter/john/12345")
}

func BenchmarkOneRouteJSON(B *testing.B) {
	app := New()
	data := struct {
		Status string `json:"status"`
	}{"ok"}
	app.GET("/json", func(c *Context) IResponse {
		return JsonResponse(200, data)
	})
	runRequest(B, app, "GET", "/json")
}

var htmlContentType = []string{"text/html; charset=utf-8"}

func BenchmarkOneRouteHTML(B *testing.B) {
	app := New()
	app.TemplatingManager().Renderer("html").LoadTemplateGlob("index", "<html><body><h1>{{.}}</h1></body></html>")
	app.GET("/html", func(c *Context) IResponse {
		return c.Renderer("html").Response(200, "index", "hola")
	})
	runRequest(B, app, "GET", "/html")
}

func BenchmarkOneRouteSet(B *testing.B) {
	app := New()
	app.GET("/ping", func(c *Context) IResponse {
		c.Set("key", "value")
		return &Response{Status: 200}
	})
	runRequest(B, app, "GET", "/ping")
}

func BenchmarkOneRouteString(B *testing.B) {
	app := New()
	app.GET("/text", func(c *Context) IResponse {
		return &Response{Status: 200, Bytes: []byte("this is a plain text")}
	})
	runRequest(B, app, "GET", "/text")
}

func BenchmarkManyRoutesFist(B *testing.B) {
	app := New()
	app.Any("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/ping")
}

func BenchmarkManyRoutesLast(B *testing.B) {
	app := New()
	app.Any("/ping", func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "OPTIONS", "/ping")
}

func Benchmark404(B *testing.B) {
	app := New()
	app.Any("/something", func(c *Context) IResponse { return &Response{Status: 200} })
	app.NoRoute(func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/ping")
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

	app.NoRoute(func(c *Context) IResponse { return &Response{Status: 200} })
	runRequest(B, app, "GET", "/viewfake")
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

func runRequest(B *testing.B, r *Application, method, path string) {
	// create fake request
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}
	w := newMockWriter()
	B.ReportAllocs()
	B.ResetTimer()
	for i := 0; i < B.N; i++ {
		r.ServeHTTP(w, req)
	}
}
