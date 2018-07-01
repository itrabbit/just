package just

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"mime/multipart"
	"net/url"
)

var (
	ErrSerializeOperationsDisabled = errors.New("serialize operations disabled")
)

// Base JSON serializer.
type JsonSerializer struct {
	Ch string // Charset for generate Content-Type header.
}

func (JsonSerializer) Name() string {
	return "json"
}

func (s JsonSerializer) Charset() string {
	return s.Ch
}

func (s JsonSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Ch) > 0 {
		return "application/json; charset=" + s.Ch
	}
	return "application/json"
}

func (JsonSerializer) Serialize(v interface{}) ([]byte, error) {
	if input, ok := v.(ISerializeInput); ok {
		v = input.Data()
	}
	if IsDebug() {
		return json.MarshalIndent(v, "", "    ")
	}
	return json.Marshal(v)
}

func (JsonSerializer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s JsonSerializer) Response(status int, data interface{}) IResponse {
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

// Base XML serializer.
type XmlSerializer struct {
	Ch string // Charset for generate Content-Type header.
}

func (XmlSerializer) Name() string {
	return "xml"
}

func (s XmlSerializer) Charset() string {
	return s.Ch
}

func (s XmlSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Ch) > 0 {
		return "application/xml; charset=" + s.Ch
	}
	return "application/xml"
}

func (XmlSerializer) Serialize(v interface{}) ([]byte, error) {
	if input, ok := v.(ISerializeInput); ok {
		v = input.Data()
	}
	if IsDebug() {
		return xml.MarshalIndent(v, "", "    ")
	}
	return xml.Marshal(v)
}

func (XmlSerializer) Deserialize(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (s XmlSerializer) Response(status int, data interface{}) IResponse {
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

const (
	defaultMaxMultipartSize = 32 << 20
	defaultMaxUrlSize       = 10 << 20
)

// Form serializer (form-data, x-www-form-urlencoded).
type FormSerializer struct {
	Ch              string
	OnlyDeserialize bool
}

func (FormSerializer) Name() string {
	return "form"
}

func (s FormSerializer) Charset() string {
	return s.Ch
}

func (s FormSerializer) DefaultContentType(withCharset bool) string {
	if withCharset && len(s.Ch) > 0 {
		return "application/x-www-form-urlencoded; charset=" + s.Ch
	}
	return "application/x-www-form-urlencoded"
}

func (s FormSerializer) Serialize(v interface{}) ([]byte, error) {
	if s.OnlyDeserialize {
		return nil, ErrSerializeOperationsDisabled
	}
	if input, ok := v.(ISerializeInput); ok {
		v = input.Data()
	}
	return marshalUrlValues(v)
}

func (FormSerializer) Deserialize(data []byte, v interface{}) error {
	if len(data) <= defaultMaxUrlSize && bytes.Index(data, []byte("\n")) < 0 {
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}
		return mapForm(values, nil, v)
	}
	if end := bytes.LastIndex(data, []byte("--")); end > 0 {
		if start := bytes.LastIndex(data[:end], []byte("\n--")); start > 0 && end > start {
			if boundary := string(data[start+3 : end]); len(boundary) > 0 {
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

func (s FormSerializer) Response(status int, data interface{}) IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return &Response{
			Status:  500,
			Bytes:   []byte("U500. Error serialize data to form urlencoded\r\n" + err.Error()),
			Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		}
	}
	return &Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}
