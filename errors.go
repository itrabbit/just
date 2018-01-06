package just

import "encoding/xml"

// Reason struct for error.
type ErrorCause struct {
	XMLName     xml.Name `json:"-" xml:"cause"`
	Path        string   `json:"path,omitempty" xml:"path,attr,omitempty"`
	Target      string   `json:"target,omitempty" xml:"target,attr,omitempty"`
	Description string   `json:"desc,omitempty" xml:"desc,omitempty"`
}

// Error struct.
type Error struct {
	XMLName  xml.Name     `json:"-" xml:"error"`
	Code     string       `json:"code,omitempty" xml:"code,attr,omitempty"`
	Message  string       `json:"msg" xml:"msg"`
	Causes   []ErrorCause `json:"causes,omitempty" xml:"causes,omitempty"`
	Metadata H            `json:"metadata,omitempty" xml:"metadata,omitempty"`
}

// Text error.
func (e *Error) Error() string {
	return e.Message
}

// Adds cause of errors.
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

// Set metadata to error.
func (e *Error) SetMetadata(meta H) *Error {
	e.Metadata = meta
	return e
}

// Method of quickly creating errors.
func NewError(code, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}
