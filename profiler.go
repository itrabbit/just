package just

import "net/http"

// Profiler interface.
type IProfiler interface {
	OnStartRequest(*http.Request)            // Event start processing HTTP request.
	OnSelectRoute(*http.Request, IRouteInfo) // Event select route for request.
	OnWriteResponseData([]byte)              // Event write data to response.
	OnWriteResponseHeader(int, http.Header)  // Event write headers to response.
	Info(...interface{})                     // Send info message to profiler.
	Error(...interface{})                    // Send error message to profiler.
	Warning(...interface{})                  // Send warning message to profiler.
	Debug(...interface{})                    // Send debug message to profiler.
}

type profiledResponseWriter struct {
	writer   http.ResponseWriter
	profiler IProfiler
}

func (w *profiledResponseWriter) Write(data []byte) (int, error) {
	defer w.profiler.OnWriteResponseData(data)
	return w.writer.Write(data)
}

func (w *profiledResponseWriter) WriteHeader(status int) {
	defer w.profiler.OnWriteResponseHeader(status, w.writer.Header())
	w.writer.WriteHeader(status)
}

func (w *profiledResponseWriter) Header() http.Header {
	return w.writer.Header()
}
