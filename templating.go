package just

import (
	"bytes"
	"errors"
	"html/template"
	"sync"
)

// Errors
var (
	ErrEmptyTemplates = errors.New("have hot templates")
)

// Interface for rendering templates.
type IRenderer interface {
	DefaultContentType(withCharset bool) string
	LoadTemplateFiles(fileNames ...string) error
	LoadTemplateGlob(name, pattern string) error
	AddFunc(name string, i interface{}) IRenderer
	RemoveFunc(name string) IRenderer
	Render(name string, data interface{}) ([]byte, error)
	Response(status int, name string, data interface{}) IResponse
}

// Interface Template Manager.
type ITemplatingManager interface {
	SetRenderer(name string, r IRenderer)
	Renderer(name string) IRenderer
}

type templatingManager struct {
	sync.RWMutex
	renderers map[string]IRenderer
}

func (t *templatingManager) SetRenderer(name string, r IRenderer) {
	t.RLock()
	defer t.RUnlock()

	if t.renderers == nil {
		t.renderers = make(map[string]IRenderer)
	}
	t.renderers[name] = r
}

func (t *templatingManager) Renderer(name string) IRenderer {
	t.Lock()
	defer t.Unlock()

	if t.renderers != nil {
		if r, ok := t.renderers[name]; ok {
			return r
		}
	}
	return nil
}

// HTML template engine.
type HTMLRenderer struct {
	Charset  string
	template *template.Template
	funcMap  map[string]interface{}
}

// Content Type by default for HTMLRenderer.
func (r *HTMLRenderer) DefaultContentType(withCharset bool) string {
	if withCharset && len(r.Charset) > 0 {
		return "text/html; charset=" + r.Charset
	}
	return "text/html"
}

// Load HTML template files.
func (r *HTMLRenderer) LoadTemplateFiles(fileNames ...string) error {
	if r.template == nil {
		t, err := template.ParseFiles(fileNames...)
		if err != nil {
			return err
		}
		r.template = t
		if r.funcMap != nil {
			r.template = r.template.Funcs(r.funcMap)
		}
		return nil
	}
	t, err := r.template.ParseFiles(fileNames...)
	if err != nil {
		return err
	}
	r.template = t
	return nil
}

// Load HTML template glob.
func (r *HTMLRenderer) LoadTemplateGlob(name, pattern string) error {
	if r.template == nil {
		t, err := template.New(name).ParseGlob(pattern)
		if err != nil {
			return err
		}
		r.template = t
		if r.funcMap != nil {
			r.template = r.template.Funcs(r.funcMap)
		}
		return nil
	}
	t, err := r.template.New(name).ParseGlob(pattern)
	if err != nil {
		return err
	}
	r.template = t
	return nil
}

// Add Func to HTML templates.
func (r *HTMLRenderer) AddFunc(name string, i interface{}) IRenderer {
	if r.funcMap == nil {
		r.funcMap = make(map[string]interface{})
	}
	r.funcMap[name] = i
	if r.template != nil {
		r.template = r.template.Funcs(r.funcMap)
	}
	return r
}

// Remove Func from HTML templates.
func (r *HTMLRenderer) RemoveFunc(name string) IRenderer {
	if r.funcMap != nil {
		delete(r.funcMap, name)
		if r.template != nil {
			r.template = r.template.Funcs(r.funcMap)
		}
	}
	return r
}

// Render HTML to bytes.
func (r *HTMLRenderer) Render(name string, data interface{}) ([]byte, error) {
	if r.template == nil {
		return nil, ErrEmptyTemplates
	}
	var buffer bytes.Buffer
	if err := r.template.ExecuteTemplate(&buffer, name, data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Response from HTML template.
func (r *HTMLRenderer) Response(status int, name string, data interface{}) IResponse {
	if b, err := r.Render(name, data); err == nil {
		return &Response{
			Status: status,
			Bytes:  b,
			Headers: map[string]string{
				ContentTypeHeaderKey: r.DefaultContentType(true),
			},
		}
	}
	return nil
}
