package finalizer

import (
	"github.com/itrabbit/just"
)

type input struct {
	v      interface{} `json:"-" xml:"-"` // Данные
	groups []string    `json:"-" xml:"-"` // Группы фильтрации
}

func (i input) Data() interface{} {
	return i.v
}

func (i input) Options() interface{} {
	return i.groups
}

func Input(v interface{}, groups ...string) interface{} {
	return &input{v: v, groups: groups}
}

type JsonSerializer struct {
	just.JsonSerializer
}

func NewJsonSerializer(charset string) just.ISerializer {
	return &JsonSerializer{just.JsonSerializer{Charset: charset}}
}

func (s JsonSerializer) Serialize(v interface{}) ([]byte, error) {
	if in, ok := v.(just.ISerializeInput); ok {
		if groups, ok := in.Options().([]string); ok && len(groups) > 0 {
			return s.JsonSerializer.Serialize(Finalize(s.Name(), in.Data(), groups...))
		}
		return s.JsonSerializer.Serialize(Finalize(s.Name(), in.Data()))
	}
	return s.JsonSerializer.Serialize(Finalize(s.Name(), v))
}

func (s JsonSerializer) Response(status int, data interface{}) just.IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return just.JsonResponse(500, just.NewError("U500", "Error serialize data to JSON").SetMetadata(just.H{"error": err.Error()}))
	}
	return &just.Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}

type XmlSerializer struct {
	just.XmlSerializer
}

func NewXmlSerializer(charset string) just.ISerializer {
	return &XmlSerializer{just.XmlSerializer{Charset: charset}}
}

func (s XmlSerializer) Serialize(v interface{}) ([]byte, error) {
	if in, ok := v.(input); ok {
		if len(in.groups) > 0 {
			return s.XmlSerializer.Serialize(Finalize(s.Name(), in.v, in.groups...))
		}
		return s.XmlSerializer.Serialize(Finalize(s.Name(), in.v))
	}
	if in, ok := v.(*input); ok {
		if len(in.groups) > 0 {
			return s.XmlSerializer.Serialize(Finalize(s.Name(), in.v, in.groups...))
		}
		return s.XmlSerializer.Serialize(Finalize(s.Name(), in.v))
	}
	return s.XmlSerializer.Serialize(Finalize(s.Name(), v))
}

func (s XmlSerializer) Response(status int, data interface{}) just.IResponse {
	b, err := s.Serialize(data)
	if err != nil {
		return just.XmlResponse(500, just.NewError("U500", "Error serialize data to XML").SetMetadata(just.H{"error": err.Error()}))
	}
	return &just.Response{
		Status:  status,
		Bytes:   b,
		Headers: map[string]string{"Content-Type": s.DefaultContentType(true)},
	}
}
