package just

import "encoding/xml"

// ErrorCause причина позникновения ошибки
type ErrorCause struct {
	XMLName     xml.Name `xml:"cause"`
	Target      string   `json:"target,omitempty" xml:"target,attr,omitempty"`
	Description string   `json:"desc,omitempty" xml:"desc,omitempty"`
}

// Error ошибка
type Error struct {
	XMLName xml.Name     `xml:"error"`
	Code    string       `json:"code,omitempty" xml:"code,attr,omitempty"`
	Message string       `json:"msg" xml:"msg"`
	Causes  []ErrorCause `json:"causes,omitempty" xml:"causes,omitempty"`
}

// Error описание ошибки для интерфейса ошибок в Go
func (e *Error) Error() string {
	return e.Message
}

// AddCause добавляет причину возникновения ошибки
func (e *Error) AddCause(target, desc string) *Error {
	if e.Causes == nil {
		e.Causes = make([]ErrorCause, 0)
	}
	e.Causes = append(e.Causes, ErrorCause{
		Target:      target,
		Description: desc,
	})
	return e
}

// NewError метод быстрого создания ошибки
func NewError(code, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}
