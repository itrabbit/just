package just

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Context - класс контекста данных запроса / ответа
type Context struct {
	// Приватные параметры
	app         IApplication      // Приложение
	routeInfo   IRouteInfo        // Данные текущего роута
	routeParams map[string]string // Параметры роутинга
	handleIndex int               // Индекс текущего обработчика

	// Открытые параметры
	Meta                map[string]interface{} // Метаданные
	Request             *http.Request          // HTTP запрос
	IsFrozenRequestBody bool                   // Замороженное тело запроса (default true)
}

// Context::reset - сбрасываем контекст
func (c *Context) reset() *Context {
	c.Request, c.routeInfo, c.routeParams, c.Meta, c.handleIndex = nil, nil, nil, nil, -1
	c.IsFrozenRequestBody = true
	return c
}

// Context::resetRoute - сбрасываем данные о текущем роуте
func (c *Context) resetRoute(info IRouteInfo, params map[string]string) *Context {
	c.routeInfo, c.routeParams, c.handleIndex = info, params, -1
	return c
}

// Context::RouteBasePath - текущий путь роута
func (c *Context) RouteBasePath() string {
	if c.routeInfo != nil {
		return c.routeInfo.BasePath()
	}
	return ""
}

// Context::Trans - перевод строки с учетом параметров траслятора
func (c *Context) Trans(message string, vars ...interface{}) string {
	if translator := c.app.Translator(); translator != nil {
		if i, ok := c.Get("locale"); ok {
			if locale, valid := i.(string); valid {
				return translator.Trans(locale, message, vars...)
			}
		}
	}
	if len(vars) > 0 {
		return fmt.Sprintf(message, vars...)
	}
	return message
}

// Context::Renderer - отрисовщик по имени
func (c *Context) Renderer(name string) IRenderer {
	return c.app.TemplatingManager().Renderer(name)
}

// Context::DetectedSerializerName - определенное имя сериализатора в запросе
func (c *Context) DetectedSerializerName() string {
	var name string
	if v, ok := c.Query("_format"); ok {
		name = v
	} else if v, ok := c.RequestHeader("FORMAT"); ok {
		name = v
	} else if v, ok := c.PostForm("_format"); ok {
		name = v
	}
	if len(name) > 1 {
		if s := c.app.SerializerManager().Serializer(name, false); s != nil {
			return name
		}
	}
	if defName, ok := c.app.SerializerManager().DefaultName(); ok {
		return defName
	}
	return "json"
}

// Context::Serializer - получить сериализатор по имени или типу контента
func (c *Context) Serializer(names ...string) ISerializer {
	m := c.app.SerializerManager()
	if len(names) > 0 {
		for _, n := range names {
			if n == "default" {
				if def, ok := m.DefaultName(); ok {
					return m.Serializer(def, false)
				}
			}
			if strings.Index(n, "/") > 0 {
				if s := m.Serializer(n, true); s != nil {
					return s
				}
			} else if s := m.Serializer(n, false); s != nil {
				return s
			}
		}
	}
	return m.Serializer(c.DetectedSerializerName(), false)
}

// Context::Bind - десериализация контента запроса в объект
func (c *Context) Bind(ptr interface{}) error {
	if c.Request == nil {
		return errors.New("Empty request")
	}
	if (c.Request.Method == "GET" || c.Request.Method == "DELETE") && c.Request.URL != nil {
		s := c.app.SerializerManager().Serializer("application/x-www-form-urlencoded", true)
		if s == nil {
			return errors.New("Not find Serializer for url query (application/x-www-form-urlencoded)")
		}
		return s.Deserialize([]byte(c.Request.URL.RawQuery), ptr)
	}
	if c.Request.Body == nil {
		return errors.New("Empty request body")
	}
	s := c.app.SerializerManager().Serializer(c.ContentType(), true)
	if s == nil {
		return errors.New("Not find Serializer for " + c.ContentType())
	}
	b, err := ioutil.ReadAll(c.Request.Body)
	defer c.ResetBodyReaderPosition()
	if err != nil {
		return err
	}
	return s.Deserialize(b, ptr)
}

// Context::IsValid - валидация контекста
func (c *Context) IsValid() bool {
	return c.app != nil && c.Request != nil
}

// Context::nextHandler - переходим к выпонение следующего handler
func (c *Context) nextHandler() (IResponse, bool) {
	if c.routeInfo != nil {
		c.handleIndex++
		if handler, ok := c.routeInfo.HandlerByIndex(c.handleIndex); ok {
			if handler != nil {
				return handler(c), true
			}
			return c.nextHandler()
		}
	}
	return nil, false
}

// Context::Next - переходим к выпонение handler
func (c *Context) Next() IResponse {
	if res, ok := c.nextHandler(); ok {
		return res
	}
	return nil
}

/**
 * Метода для работа с параметрами роутинга
 */

// Context::Param - получаем параметр из роутера
func (c *Context) Param(name string) (value string, ok bool) {
	if c.routeParams != nil {
		value, ok = c.routeParams[name]
	}
	return
}

// Context::ParamBool - получаем bool параметр из роутера
func (c *Context) ParamBool(name string) (value bool, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseBool(strings.ToLower(str))
		ok = (err == nil)
	}
	return
}

// Context::ParamInt - получаем int параметр из роутера
func (c *Context) ParamInt(name string) (value int64, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseInt(str, 10, 64)
		ok = (err == nil)
	}
	return
}

// Context::ParamFloat - получаем int параметр из роутера
func (c *Context) ParamFloat(name string) (value float64, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseFloat(str, 64)
		ok = (err == nil)
	}
	return
}

// Context::ParamDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) ParamDef(name string, def string) string {
	if value, ok := c.Param(name); ok {
		return value
	}
	return def
}

// Context::ParamBoolDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) ParamBoolDef(name string, def bool) bool {
	if value, ok := c.ParamBool(name); ok {
		return value
	}
	return def
}

// Context::ParamIntDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) ParamIntDef(name string, def int64) int64 {
	if value, ok := c.ParamInt(name); ok {
		return value
	}
	return def
}

// Context::ParamFloatDef - получаем параметр из роутера с учетом значения по умолчанию
func (c *Context) ParamFloatDef(name string, def float64) float64 {
	if value, ok := c.ParamFloat(name); ok {
		return value
	}
	return def
}

func (c *Context) MustParam(name string) string {
	return c.ParamDef(name, "")
}

func (c *Context) MustParamBool(name string) bool {
	return c.ParamBoolDef(name, false)
}

func (c *Context) MustParamInt(name string) int64 {
	return c.ParamIntDef(name, 0)
}

func (c *Context) MustParamFloat(name string) float64 {
	return c.ParamFloatDef(name, 0)
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

func (c *Context) MustGet(key string) interface{} {
	return c.GetDef(key, nil)
}

/**
 * Методы для работы с условиями Query Url
 */

func (c *Context) Query(key string) (string, bool) {
	if values, ok := c.QueryArray(key); ok && len(values) > 0 {
		return values[0], true
	}
	return "", false
}

func (c *Context) QueryDef(key, def string) string {
	if value, ok := c.Query(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQuery(key string) string {
	return c.QueryDef(key, "")
}

func (c *Context) QueryBool(key string) (bool, bool) {
	if str, ok := c.Query(key); ok {
		if b, err := strconv.ParseBool(strings.ToLower(str)); err == nil {
			return b, true
		}
	}
	return false, false
}

func (c *Context) QueryBoolDef(key string, def bool) bool {
	if value, ok := c.QueryBool(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryBool(key string) bool {
	return c.QueryBoolDef(key, false)
}

func (c *Context) QueryInt(key string) (int64, bool) {
	if str, ok := c.Query(key); ok {
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func (c *Context) QueryIntDef(key string, def int64) int64 {
	if value, ok := c.QueryInt(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryInt(key string) int64 {
	return c.QueryIntDef(key, 0)
}

func (c *Context) QueryFloat(key string) (float64, bool) {
	if str, ok := c.Query(key); ok {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (c *Context) QueryFloatDef(key string, def float64) float64 {
	if value, ok := c.QueryFloat(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryFloat(key string) float64 {
	return c.QueryFloatDef(key, 0)
}

func (c *Context) QueryArray(key string) ([]string, bool) {
	if values, ok := c.Request.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

func (c *Context) QueryArrayDef(key string, def []string) []string {
	if values, ok := c.QueryArray(key); ok {
		return values
	}
	return def
}

func (c *Context) MustQueryArray(key string) []string {
	return c.QueryArrayDef(key, make([]string, 0))
}

/**
 * Методы для работы с данными в Body / POST
 */

func (c *Context) PostForm(key string) (string, bool) {
	if values, ok := c.PostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormDef(key, def string) string {
	if value, ok := c.PostForm(key); ok {
		return value
	}
	return def
}

func (c *Context) MustPostForm(key string) string {
	return c.PostFormDef(key, "")
}

func (c *Context) PostFormBool(key string) (bool, bool) {
	if str, ok := c.PostForm(key); ok {
		if b, err := strconv.ParseBool(str); err == nil {
			return b, true
		}
	}
	return false, false
}

func (c *Context) PostFormBoolDef(key string, def bool) bool {
	if value, ok := c.PostFormBool(key); ok {
		return value
	}
	return def
}

func (c *Context) MustPostFormBool(key string) bool {
	return c.PostFormBoolDef(key, false)
}

func (c *Context) PostFormInt(key string) (int64, bool) {
	if str, ok := c.PostForm(key); ok {
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func (c *Context) PostFormIntDef(key string, def int64) int64 {
	if value, ok := c.PostFormInt(key); ok {
		return value
	}
	return def
}

func (c *Context) MustPostFormInt(key string) int64 {
	return c.PostFormIntDef(key, 0)
}

func (c *Context) PostFormFloat(key string) (float64, bool) {
	if str, ok := c.PostForm(key); ok {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (c *Context) PostFormFloatDef(key string, def float64) float64 {
	if value, ok := c.PostFormFloat(key); ok {
		return value
	}
	return def
}

func (c *Context) MustPostFormFloat(key string) float64 {
	return c.PostFormFloatDef(key, 0)
}

func (c *Context) PostFormArray(key string) ([]string, bool) {
	c.Request.ParseForm()
	c.Request.ParseMultipartForm(32 << 20) // 32 MB
	if c.IsFrozenRequestBody {
		c.ResetBodyReaderPosition()
	}
	if values := c.Request.PostForm[key]; len(values) > 0 {
		return values, true
	}
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		if values := c.Request.MultipartForm.Value[key]; len(values) > 0 {
			return values, true
		}
	}
	return []string{}, false
}

func (c *Context) PostFormArrayDef(key string, def []string) []string {
	if values, ok := c.PostFormArray(key); ok {
		return values
	}
	return def
}

func (c *Context) MustPostFormArray(key string) []string {
	return c.PostFormArrayDef(key, make([]string, 0))
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.Request.FormFile(name)
	if c.IsFrozenRequestBody {
		c.ResetBodyReaderPosition()
	}
	return fh, err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMaxMultipartSize)
	if c.IsFrozenRequestBody {
		c.ResetBodyReaderPosition()
	}
	return c.Request.MultipartForm, err
}

/**
 * Методы для работы с данными запроса
 */

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

func (c *Context) CookieDef(name string, def string) string {
	if cookie, err := c.Cookie(name); err == nil {
		return cookie
	}
	return def
}

func (c *Context) MustCookie(name string) string {
	return c.CookieDef(name, "")
}

func (c *Context) ClientIP(forwarded bool) string {
	if forwarded {
		clientIP := strings.TrimSpace(c.MustRequestHeader("X-Real-Ip"))
		if len(clientIP) > 0 {
			return clientIP
		}
		clientIP = c.MustRequestHeader("X-Forwarded-For")
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) > 0 {
			return clientIP
		}
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func (c *Context) UserAgent() string {
	return c.MustRequestHeader("User-Agent")
}

func (c *Context) ContentType() string {
	contentType := c.MustRequestHeader("Content-Type")
	for i, ch := range contentType {
		if ch == ' ' || ch == ';' {
			contentType = strings.TrimSpace(contentType[:i])
			break
		}
	}
	if len(contentType) < 1 {
		contentType = "text/plain"
	}
	return contentType
}

func (c *Context) RequestHeader(key string) (string, bool) {
	if values, ok := c.Request.Header[key]; len(values) > 0 && ok {
		return values[0], true
	}
	return "", false
}

func (c *Context) RequestHeaderDef(key string, def string) string {
	if value, ok := c.RequestHeader(key); ok {
		return value
	}
	return def
}

func (c *Context) MustRequestHeader(key string) string {
	return c.RequestHeaderDef(key, "")
}

func (c *Context) ResetBodyReaderPosition() error {
	_, err := topSeekReader(c.Request.Body, true)
	return err
}
