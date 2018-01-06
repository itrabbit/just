package just

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Response interface.
type IResponse interface {
	GetData() []byte
	GetStatus() int
	GetHeaders() map[string]string
	GetStreamHandler() (http.HandlerFunc, bool)
}

// Base Response struct.
type Response struct {
	Status  int               // HTTP status (200, 201,...).
	Bytes   []byte            // Data bytes.
	Headers map[string]string // Response headers.
	Stream  http.HandlerFunc  // Stream method, to support standard HTTP package, as well as to work with WS.
}

func (r *Response) GetStreamHandler() (http.HandlerFunc, bool) {
	if r.Stream != nil && (r.Bytes == nil || len(r.Bytes) < 1) {
		return r.Stream, true
	}
	return nil, false
}

func (r *Response) GetStatus() int {
	return r.Status
}

func (r *Response) GetData() []byte {
	return r.Bytes
}

func (r *Response) GetHeaders() map[string]string {
	return r.Headers
}

// Create a stream response.
func StreamResponse(handler http.HandlerFunc) IResponse {
	return &Response{Bytes: nil, Status: -1, Headers: nil, Stream: handler}
}

// Create a JSON response.
func JsonResponse(status int, v interface{}) IResponse {
	b, err := json.Marshal(v)
	if err != nil {
		return &Response{
			Bytes:   []byte(err.Error()),
			Status:  500,
			Headers: map[string]string{"Content-Type": "plain/text; charset=utf-8"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
	}
}

// Create a redirect response (Use _StrongRedirect in header to set location)
func RedirectResponse(status int, location string) IResponse {
	if (status < 300 || status > 308) && status != 201 {
		status = 301
	}
	return &Response{Bytes: nil, Status: status, Headers: map[string]string{"_StrongRedirect": location}}
}

// Create a XML response.
func XmlResponse(status int, v interface{}) IResponse {
	b, err := xml.Marshal(v)
	if err != nil {
		return &Response{
			Bytes:   []byte(err.Error()),
			Status:  500,
			Headers: map[string]string{"Content-Type": "plain/text; charset=utf-8"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/xml; charset=utf-8"},
	}
}

// Create a response in the form of a local file.
func FileResponse(filePath string) IResponse {
	return &Response{Bytes: nil, Status: -1, Headers: map[string]string{"_FilePath": filePath}}
}
