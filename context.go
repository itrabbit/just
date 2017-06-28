package just

import "net/http"

// Context - класс контекста данных запроса / ответа
type Context struct {
	app     *Application  // Приложение
	request *http.Request // HTTP запрос
}

// Context::reset - сбрасываем контекст
func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.request = req
}

// Context::Request - получаем данные запроса
func (c *Context) Request() *http.Request {
	return c.request
}
