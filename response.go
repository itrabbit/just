package just

import (
	"encoding/json"
	"encoding/xml"
)

type IResponse interface {
	GetData() []byte
	GetStatus() int
	GetHeaders() map[string]string
}

type Response struct {
	Status  int
	Bytes   []byte
	Headers map[string]string
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

// NewJsonResponse создание ответа в формате JSON
func NewJsonResponse(status int, v interface{}) IResponse {
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

// NewXmlResponse создание ответа в формате xml
func NewXmlResponse(status int, v interface{}) IResponse {
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
