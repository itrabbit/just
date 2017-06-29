package just

import (
	"net/http"
)

// Context - класс контекста данных запроса / ответа
type Context struct {
	app         *Application      // Приложение
	request     *http.Request     // HTTP запрос
	routeInfo   IRouteInfo        // Данные текущего роута
	routeParams map[string]string // Параметры роутинга
	handleIndex int               // Индекс текущего обработчика
}

// Context::reset - сбрасываем контекст
func (c *Context) reset(req *http.Request) {
	c.request, c.routeInfo, c.routeParams, c.handleIndex = req, nil, nil, -1
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

// Context::GetParam - получаем параметр из роутера
func (c *Context) GetParam(name string) (string, bool) {
	if c.routeParams != nil {
		if value, ok := c.routeParams[name]; ok {
			return value, ok
		}
	}
	return "", false
}

// Context::GetParamDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) GetParamDef(name string, def string) string {
	if value, ok := c.GetParam(name); ok {
		return value
	}
	return def
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
