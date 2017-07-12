package just

import "encoding/xml"

// ErrorCause причина позникновения ошибки
type ErrorCause struct {
	XMLName     xml.Name `json:"-" xml:"cause"`
	Path        string   `json:"path,omitempty" xml:"path,attr,omitempty"`
	Target      string   `json:"target,omitempty" xml:"target,attr,omitempty"`
	Description string   `json:"desc,omitempty" xml:"desc,omitempty"`
}

// Error ошибка
type Error struct {
	XMLName  xml.Name     `json:"-" xml:"error"`
	Code     string       `json:"code,omitempty" xml:"code,attr,omitempty"`
	Message  string       `json:"msg" xml:"msg"`
	Causes   []ErrorCause `json:"causes,omitempty" xml:"causes,omitempty"`
	Metadata H            `json:"metadata,omitempty" xml:"metadata,omitempty"`
}

// Error описание ошибки для интерфейса ошибок в Go
func (e *Error) Error() string {
	return e.Message
}

// AddCause добавляет причину возникновения ошибки
func (e *Error) AddCause(target, path, desc string) *Error {
	if e.Causes == nil {
		e.Causes = make([]ErrorCause, 0)
	}
	e.Causes = append(e.Causes, ErrorCause{
		Path:        path,
		Target:      target,
		Description: desc,
	})
	return e
}

func (e *Error) SetMetadata(meta H) *Error {
	e.Metadata = meta
	return e
}

// NewError метод быстрого создания ошибки
func NewError(code, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}
