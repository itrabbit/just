package just

import "net/http"

type IProfiler interface {
	StartRequest(*http.Request)
	SelectRoute(*http.Request, IRouteInfo)
	WriteResponseData([]byte)
	WriteResponseHeader(int, http.Header)
	Info(...interface{})
	Error(...interface{})
	Warning(...interface{})
	Debug(...interface{})
}

type profiledResponseWriter struct {
	writer   http.ResponseWriter
	profiler IProfiler
}

func (w *profiledResponseWriter) Write(data []byte) (int, error) {
	// Фиксация записи данных в ответ
	defer w.profiler.WriteResponseData(data)
	return w.writer.Write(data)
}

func (w *profiledResponseWriter) WriteHeader(status int) {
	defer w.profiler.WriteResponseHeader(status, w.writer.Header())
	w.writer.WriteHeader(status)
}

func (w *profiledResponseWriter) Header() http.Header {
	return w.writer.Header()
}
