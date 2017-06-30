package just

import (
	"net/http"
	"strconv"
)

// Context - класс контекста данных запроса / ответа
type Context struct {
	// Приватные параметры
	app         *Application      // Приложение
	request     *http.Request     // HTTP запрос
	routeInfo   IRouteInfo        // Данные текущего роута
	routeParams map[string]string // Параметры роутинга
	handleIndex int               // Индекс текущего обработчика

	// Открытые параметры
	Meta map[string]interface{} // Метаданные
}

// Context::reset - сбрасываем контекст
func (c *Context) reset(req *http.Request) *Context {
	c.request, c.routeInfo, c.routeParams, c.handleIndex = req, nil, nil, -1
	c.Meta = nil
	return c
}

// Context::resetRoute - сбрасываем данные о текущем роуте
func (c *Context) resetRoute(info IRouteInfo, params map[string]string) *Context {
	c.routeInfo, c.routeParams, c.handleIndex = info, params, -1
	return c
}

func (c *Context) RouteBasePath() string {
	if c.routeInfo != nil {
		return c.routeInfo.BasePath()
	}
	return ""
}

// Context::Request - получаем данные запроса
func (c *Context) Request() *http.Request {
	return c.request
}

// Context::IsValid - валидация контекста
func (c *Context) IsValid() bool {
	return c.request != nil
}

// Context::Next - переходим к выпонение handler
func (c *Context) Next() IResponse {
	c.handleIndex++
	if handler := c.routeInfo.HandlerByIndex(c.handleIndex); handler != nil {
		return handler(c)
	}
	return nil
}

/**
 * Метода для работа с параметрами роутинга
 */

// Context::GetParam - получаем параметр из роутера
func (c *Context) GetParam(name string) (value string, ok bool) {
	if c.routeParams != nil {
		value, ok = c.routeParams[name]
	}
	return
}

// Context::GetIntParam - получаем int параметр из роутера
func (c *Context) GetIntParam(name string) (value int64, ok bool) {
	if str, exist := c.GetParam(name); exist {
		var err error
		value, err = strconv.ParseInt(str, 10, 64)
		ok = (err == nil)
	}
	return
}

// Context::GetFloatParam - получаем int параметр из роутера
func (c *Context) GetFloatParam(name string) (value float64, ok bool) {
	if str, exist := c.GetParam(name); exist {
		var err error
		value, err = strconv.ParseFloat(str, 64)
		ok = (err == nil)
	}
	return
}

// Context::GetParamDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) GetParamDef(name string, def string) string {
	if value, ok := c.GetParam(name); ok {
		return value
	}
	return def
}

// Context::GetIntParamDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) GetIntParamDef(name string, def int64) int64 {
	if value, ok := c.GetIntParam(name); ok {
		return value
	}
	return def
}

// Context::GetFloatParamDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) GetFloatParamDef(name string, def float64) float64 {
	if value, ok := c.GetFloatParam(name); ok {
		return value
	}
	return def
}

/**
 * Методы для работа с метаданными
 */

// Context::Set - устанавливаем метаданные по ключу
func (c *Context) Set(key string, value interface{}) *Context {
	if c.Meta == nil {
		c.Meta = make(map[string]interface{})
	}
	c.Meta[key] = value
	return c
}

// Context::Get - получаем метаданные по ключу
func (c *Context) Get(key string) (value interface{}, ok bool) {
	if c.Meta != nil {
		value, ok = c.Meta[key]
	}
	return
}

// Context::GetDef - получаем метаданные по ключу с значением по умолчанию
func (c *Context) GetDef(key string, def interface{}) interface{} {
	if value, ok := c.Get(key); ok {
		return value
	}
	return def
}

/**
 * Методы для работы с условиями Query Url
 */

func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

func (c *Context) QueryDef(key, def string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return def
}

func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) QueryArray(key string) []string {
	values, _ := c.GetQueryArray(key)
	return values
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	req := c.Request()
	if values, ok := req.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}