package just

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	"unicode/utf8"
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

func hexEscapeNonASCII(s string) string {
	newLen := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			newLen += 3
		} else {
			newLen++
		}
	}
	if newLen == len(s) {
		return s
	}
	b := make([]byte, 0, newLen)
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			b = append(b, '%')
			b = strconv.AppendInt(b, int64(s[i]), 16)
		} else {
			b = append(b, s[i])
		}
	}
	return string(b)
}

// RedirectResponse создание редиректа
func RedirectResponse(status int, location string) IResponse {
	if (status < 300 || status > 308) && status != 201 {
		status = 301
	}
	return &Response{Bytes: nil, Status: status, Headers: map[string]string{"Location": location, "_ThisStrongRedirect": "1"}}
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
