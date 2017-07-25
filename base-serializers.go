package just

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"net/url"
	"regexp"
)

/** JSON Serializer */

type JsonSerializer struct {
	Charset string
}

func (s *JsonSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Charset) > 0 {
		return "application/json; charset=" + s.Charset
	}
	return "application/json"
}

func (JsonSerializer) Serialize(v interface{}) ([]byte, error) {
	if IsDebug() {
		return json.MarshalIndent(v, "", "    ")
	}
	return json.Marshal(v)
}

func (JsonSerializer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *JsonSerializer) Response(status int, data interface{}) IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return JsonResponse(500, NewError("U500", "Error serialize data to JSON").SetMetadata(H{"error": err.Error()}))
	}
	return &Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}

/** XML Serializer */

type XmlSerializer struct {
	Charset string
}

func (s *XmlSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Charset) > 0 {
		return "application/xml; charset=" + s.Charset
	}
	return "application/xml"
}

func (XmlSerializer) Serialize(v interface{}) ([]byte, error) {
	if IsDebug() {
		return xml.MarshalIndent(v, "", "    ")
	}
	return xml.Marshal(v)
}

func (XmlSerializer) Deserialize(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (s *XmlSerializer) Response(status int, data interface{}) IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return XmlResponse(500, NewError("U500", "Error serialize data to XML").SetMetadata(H{"error": err.Error()}))
	}
	return &Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}

/** Form Serializer */

const (
	defaultMaxMultipartSize = 32 << 20
	defaultMaxUrlSize       = 10 << 20
)

var (
	urlencodedRx = regexp.MustCompile("([A-Za-z0-9%./]+=[^\\s]+)")
)

type FormSerializer struct {
	Charset string
}

func (s *FormSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Charset) > 0 {
		return "application/x-www-form-urlencoded; charset=" + s.Charset
	}
	return "application/x-www-form-urlencoded"
}

func (FormSerializer) Serialize(v interface{}) ([]byte, error) {
	return marshalUrlValues(v)
}

func (FormSerializer) Deserialize(data []byte, v interface{}) error {
	if len(data) <= defaultMaxUrlSize && urlencodedRx.Match(data) {
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}
		return mapForm(values, nil, v)
	}
	if end := bytes.LastIndex(data, []byte("--")); end > 0 {
		if start := bytes.LastIndex(data[:end], []byte("\n--")); start > 0 && end > start {
			if boundary := string(data[start:end]); len(boundary) > 0 {
				r := multipart.NewReader(bytes.NewReader(data), boundary)
				form, err := r.ReadForm(defaultMaxMultipartSize)
				if err != nil {
					return err
				}
				return mapForm(form.Value, form.File, v)
			}
		}
	}
	return nil
}

func (s *FormSerializer) Response(status int, data interface{}) IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return &Response{
			Status:500,
			Bytes: []byte("U500. Error serialize data to form urlencoded\r\n"+err.Error()),
			Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		}
	}
	return &Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}