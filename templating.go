package just

import (
	"html/template"
	"sync"
	"errors"
	"bytes"
)

// Интерфейс отрисовки шаблонов
type IRenderer interface {
	DefaultContentType() string
	LoadTemplateFiles(filenames ...string) error
	LoadTemplateGlob(name, pattern string) error
	AddFunc(name string, i interface{}) IRenderer
	RemoveFunc(name string) IRenderer
	Render(name string, data interface{}) ([]byte, error)
	Response(status int, name string, data interface{}) IResponse
}

// templatingManager менеджер шаблонизаторов
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

/**
 * HTML шаблонизатор
 */

type HTMLRenderer struct {
	template *template.Template
	funcMap  map[string]interface{}
}

func (r *HTMLRenderer) DefaultContentType() string {
	return "text/html"
}

func (r *HTMLRenderer) LoadTemplateFiles(filenames ...string) error {
	if r.template == nil {
		t, err := template.ParseFiles(filenames...)
		if err != nil {
			return err
		}
		r.template = t
		if r.funcMap != nil {
			r.template = r.template.Funcs(r.funcMap)
		}
		return nil
	}
	t, err := r.template.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	r.template = t
	return nil
}

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
func (r *HTMLRenderer) RemoveFunc(name string) IRenderer {
	if r.funcMap != nil {
		delete(r.funcMap, name)
		if r.template != nil {
			r.template = r.template.Funcs(r.funcMap)
		}
	}
	return r
}

func (r *HTMLRenderer) Render(name string, data interface{}) ([]byte, error) {
	if r.template == nil {
		return nil, errors.New("Have hot templates")
	}
	var buffer bytes.Buffer
	if err := r.template.ExecuteTemplate(buffer, name, data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (r *HTMLRenderer) Response(status int, name string, data interface{}) IResponse {
	if b, err := r.Render(name, data); err == nil {
		return &Response{
			Status: status,
			Bytes: b,
			Headers: map[string]string{
				"Content-Type": r.DefaultContentType(),
			},
		}
	}
	return nil
}
