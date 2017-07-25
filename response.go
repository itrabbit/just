package just

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type IResponse interface {
	GetData() []byte
	GetStatus() int
	GetHeaders() map[string]string
	GetStreamHandler() (http.HandlerFunc, bool)
}

type Response struct {
	Status  int
	Bytes   []byte
	Headers map[string]string
	Stream  http.HandlerFunc
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

// StreamResponse создание потока ответа
func StreamResponse(handler http.HandlerFunc) IResponse {
	return &Response{Bytes: nil, Status: -1, Headers: nil, Stream: handler}
}

// JsonResponse создание ответа в формате JSON
func JsonResponse(status int, v interface{}) IResponse {
	b, err := json.Marshal(v)
	if err != nil {
		return &Response{
			Bytes:   []byte(err.Error()),
			Status:  500,
			Headers: map[string]string{"Content-Type": "plain/text"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/json"},
	}
}

// RedirectResponse создание жесткого редиректа
func RedirectResponse(status int, location string) IResponse {
	if (status < 300 || status > 308) && status != 201 {
		status = 301
	}
	return &Response{Bytes: nil, Status: status, Headers: map[string]string{"_StrongRedirect": location}}
}

// XmlResponse создание ответа в формате xml
func XmlResponse(status int, v interface{}) IResponse {
	b, err := xml.Marshal(v)
	if err != nil {
		return &Response{
			Bytes:   []byte(err.Error()),
			Status:  500,
			Headers: map[string]string{"Content-Type": "plain/text"},
		}
	}
	return &Response{
		Bytes:   b,
		Status:  status,
		Headers: map[string]string{"Content-Type": "application/xml"},
	}
}

// XmlResponse создание ответа в виде локального файла
func FileResponse(filePath string) IResponse {
	return &Response{Bytes: nil, Status: -1, Headers: map[string]string{"_FilePath": filePath}}
}
