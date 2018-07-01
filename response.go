package just

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

const (
	ContentTypeHeaderKey    = "Content-Type"
	StrongRedirectHeaderKey = "__StrongRedirect"
	ServeFileHeaderKey      = "__ServeFilePath"
)

// Response interface.
type IResponse interface {
	// Checkers
	HasData() bool
	HasHeaders() bool
	HasStreamHandler() bool
	// Getters
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

func (r Response) HasStreamHandler() bool {
	return r.Stream != nil
}

func (r Response) GetStreamHandler() (http.HandlerFunc, bool) {
	if r.Stream != nil {
		return r.Stream, true
	}
	return nil, false
}

func (r Response) HasData() bool {
	if r.Bytes == nil {
		return false
	}
	return len(r.Bytes) > 0
}

func (r Response) GetStatus() int {
	return r.Status
}

func (r Response) GetData() []byte {
	return r.Bytes
}

func (r Response) HasHeaders() bool {
	if r.Headers == nil {
		return false
	}
	return len(r.Headers) > 0
}

func (r *Response) GetHeaders() map[string]string {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
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
			Headers: map[string]string{ContentTypeHeaderKey: "plain/text; charset=utf-8"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{ContentTypeHeaderKey: "application/json; charset=utf-8"},
	}
}

// Create a redirect response (Use _StrongRedirect in header to set location)
func RedirectResponse(status int, location string) IResponse {
	if (status < 300 || status > 308) && status != 201 {
		status = 301
	}
	return &Response{Bytes: nil, Status: status, Headers: map[string]string{StrongRedirectHeaderKey: location}}
}

// Create a XML response.
func XmlResponse(status int, v interface{}) IResponse {
	b, err := xml.Marshal(v)
	if err != nil {
		return &Response{
			Bytes:   []byte(err.Error()),
			Status:  500,
			Headers: map[string]string{ContentTypeHeaderKey: "plain/text; charset=utf-8"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{ContentTypeHeaderKey: "application/xml; charset=utf-8"},
	}
}

// Create a response in the form of a local file.
func FileResponse(filePath string) IResponse {
	return &Response{Bytes: nil, Status: -1, Headers: map[string]string{ServeFileHeaderKey: filePath}}
}
