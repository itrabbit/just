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
	"time"
)

// Errors
var (
	ErrEmptyRequest          = errors.New("empty request")
	ErrEmptyRequestBody      = errors.New("empty request body")
	ErrNotFoundSerializer    = errors.New("not find serializer for content type in your request")
	ErrNotFoundUrlSerializer = errors.New("not find serializer for url query (application/x-www-form-urlencoded)")
)

// Request Context struct.
type Context struct {
	// Private props.
	app            IApplication      // Application.
	routeInfo      IRouteInfo        // Current route info.
	routeParams    map[string]string // Current route params.
	handleIndex    int               // Current handler index.
	isLocalRequest bool

	// Public props.
	Meta                map[string]interface{} // Metadata.
	Request             *http.Request          // HTTP request.
	IsFrozenRequestBody bool                   // Use frozen request body (default true).
}

func (c *Context) reset() *Context {
	c.Request, c.routeInfo, c.routeParams, c.Meta, c.handleIndex = nil, nil, nil, nil, -1
	c.IsFrozenRequestBody = true
	return c
}

func (c *Context) resetRoute(info IRouteInfo, params map[string]string) *Context {
	c.routeInfo, c.routeParams, c.handleIndex = info, params, -1
	return c
}

// Current route path.
func (c *Context) RouteBasePath() string {
	if c.routeInfo != nil {
		return c.routeInfo.BasePath()
	}
	return ""
}

// Translate text by current language.
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

// Short name for translate text method.
func (c *Context) Tr(message string, vars ...interface{}) string {
	return c.Trans(message, vars...)
}

// Current template renderer.
func (c *Context) Renderer(name string) IRenderer {
	return c.app.TemplatingManager().Renderer(name)
}

// Short name for current renderer.
func (c *Context) R(name string) IRenderer {
	return c.Renderer(name)
}

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

// Get serializer by name or content type (without names - default serializer, or auto detected serializer).
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

// Short name for serializer method
func (c *Context) S(names ...string) ISerializer {
	return c.Serializer(names...)
}

// DeSerializing body or query to object
func (c *Context) Bind(ptr interface{}) error {
	if c.Request == nil {
		return ErrEmptyRequest
	}
	if (c.Request.Method == "GET" || c.Request.Method == "DELETE") && c.Request.URL != nil {
		s := c.app.SerializerManager().Serializer("application/x-www-form-urlencoded", true)
		if s == nil {
			return ErrNotFoundUrlSerializer
		}
		return s.Deserialize([]byte(c.Request.URL.RawQuery), ptr)
	}
	if c.Request.Body == nil {
		return ErrEmptyRequestBody
	}
	s := c.app.SerializerManager().Serializer(c.ContentType(), true)
	if s == nil {
		return ErrNotFoundSerializer
	}
	b, err := ioutil.ReadAll(c.Request.Body)
	defer c.ResetBodyReaderPosition()
	if err != nil {
		return err
	}
	return s.Deserialize(b, ptr)
}

func (c *Context) IsValid() bool {
	return c.app != nil && c.Request != nil
}

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

// Turn to elective next handler
func (c *Context) Next() IResponse {
	if res, ok := c.nextHandler(); ok {
		return res
	}
	return nil
}

// Get route param by name.
func (c *Context) Param(name string) (value string, ok bool) {
	if c.routeParams != nil {
		value, ok = c.routeParams[name]
	}
	return
}

// Get bool route param by name.
func (c *Context) ParamBool(name string) (value bool, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseBool(strings.ToLower(str))
		ok = err == nil
	}
	return
}

// Get integer route param by name.
func (c *Context) ParamInt(name string) (value int64, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseInt(str, 10, 64)
		ok = err == nil
	}
	return
}

// Get float route param by name.
func (c *Context) ParamFloat(name string) (value float64, ok bool) {
	if str, exist := c.Param(name); exist {
		var err error
		value, err = strconv.ParseFloat(str, 64)
		ok = err == nil
	}
	return
}

// Get route param by name with the default value.
func (c *Context) ParamDef(name string, def string) string {
	if value, ok := c.Param(name); ok {
		return value
	}
	return def
}

// Get bool route param by name with the default value.
func (c *Context) ParamBoolDef(name string, def bool) bool {
	if value, ok := c.ParamBool(name); ok {
		return value
	}
	return def
}

// Get integer route param by name with the default value.
func (c *Context) ParamIntDef(name string, def int64) int64 {
	if value, ok := c.ParamInt(name); ok {
		return value
	}
	return def
}

// Get float route param by name with the default value.
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

// Set metadata by key.
func (c *Context) Set(key string, value interface{}) *Context {
	if c.Meta == nil {
		c.Meta = make(map[string]interface{})
	}
	c.Meta[key] = value
	return c
}

// Get metadata by key.
func (c *Context) Get(key string) (value interface{}, ok bool) {
	if c.Meta != nil {
		value, ok = c.Meta[key]
	}
	return
}

// Get metadata by key with default value.
func (c *Context) GetDef(key string, def interface{}) interface{} {
	if value, ok := c.Get(key); ok {
		return value
	}
	return def
}

// Get bool metadata by key.
func (c *Context) GetBool(key string) (value bool, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(bool)
		}
	}
	return
}

// Get bool metadata by key with default value.
func (c *Context) GetBoolDef(key string, def bool) bool {
	if value, ok := c.GetBool(key); ok {
		return value
	}
	return def
}

// Get string metadata by key.
func (c *Context) GetStr(key string) (value string, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(string)
		}
	}
	return
}

// Get string metadata by key with default value.
func (c *Context) GetStrDef(key string, def string) string {
	if value, ok := c.GetStr(key); ok {
		return value
	}
	return def
}

// Get duration metadata by key.
func (c *Context) GetDuration(key string) (value time.Duration, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(time.Duration)
		}
	}
	return
}

// Get duration metadata by key with default value.
func (c *Context) GetDurationDef(key string, def time.Duration) time.Duration {
	if value, ok := c.GetDuration(key); ok {
		return value
	}
	return def
}

// Get integer metadata by key.
func (c *Context) GetInt(key string) (value int64, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(int64)
		}
	}
	return
}

// Get unsigned integer metadata by key.
func (c *Context) GetUint(key string) (value uint64, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(uint64)
		}
	}
	return
}

// Get integer metadata by key with default value.
func (c *Context) GetIntDef(key string, def int64) int64 {
	if value, ok := c.GetInt(key); ok {
		return value
	}
	return def
}

// Get unsigned integer metadata by key with default value.
func (c *Context) GetUintDef(key string, def uint64) uint64 {
	if value, ok := c.GetUint(key); ok {
		return value
	}
	return def
}

// Get integer metadata by key.
func (c *Context) GetFloat(key string) (value float64, ok bool) {
	if c.Meta != nil {
		if i, has := c.Meta[key]; has && i != nil {
			value, ok = i.(float64)
		}
	}
	return
}

// Get float metadata by key with default value.
func (c *Context) GetFloatDef(key string, def float64) float64 {
	if value, ok := c.GetFloat(key); ok {
		return value
	}
	return def
}

func (c *Context) MustGet(key string) interface{} {
	return c.GetDef(key, nil)
}

// Get query param by key.
func (c *Context) Query(key string) (string, bool) {
	if values, ok := c.QueryArray(key); ok && len(values) > 0 {
		return values[0], true
	}
	return "", false
}

// Get query param by key with default value.
func (c *Context) QueryDef(key, def string) string {
	if value, ok := c.Query(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQuery(key string) string {
	return c.QueryDef(key, "")
}

// Get bool query param by key.
func (c *Context) QueryBool(key string) (bool, bool) {
	if str, ok := c.Query(key); ok {
		if b, err := strconv.ParseBool(strings.ToLower(str)); err == nil {
			return b, true
		}
	}
	return false, false
}

// Get bool query param by key with default value.
func (c *Context) QueryBoolDef(key string, def bool) bool {
	if value, ok := c.QueryBool(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryBool(key string) bool {
	return c.QueryBoolDef(key, false)
}

// Get integer query param by key.
func (c *Context) QueryInt(key string) (int64, bool) {
	if str, ok := c.Query(key); ok {
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

// Get integer query param by key with default value.
func (c *Context) QueryIntDef(key string, def int64) int64 {
	if value, ok := c.QueryInt(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryInt(key string) int64 {
	return c.QueryIntDef(key, 0)
}

// Get float query param by key.
func (c *Context) QueryFloat(key string) (float64, bool) {
	if str, ok := c.Query(key); ok {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// Get float query param by key with default value.
func (c *Context) QueryFloatDef(key string, def float64) float64 {
	if value, ok := c.QueryFloat(key); ok {
		return value
	}
	return def
}

func (c *Context) MustQueryFloat(key string) float64 {
	return c.QueryFloatDef(key, 0)
}

// Get array query param by key.
func (c *Context) QueryArray(key string) ([]string, bool) {
	if values, ok := c.Request.URL.Query()[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// Get array query param by key with default value.
func (c *Context) QueryArrayDef(key string, def []string) []string {
	if values, ok := c.QueryArray(key); ok {
		return values
	}
	return def
}

func (c *Context) MustQueryArray(key string) []string {
	return c.QueryArrayDef(key, make([]string, 0))
}

// PostForm returns the specified key from a POST urlencoded form or multipart form when it exists, otherwise it returns an empty string `("")`.
func (c *Context) PostForm(key string) (string, bool) {
	if values, ok := c.PostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostForm returns the specified key from a POST urlencoded form or multipart form with default value.
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

// Get multipart.FileHeader from MultipartForm by name.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.Request.FormFile(name)
	if c.IsFrozenRequestBody {
		c.ResetBodyReaderPosition()
	}
	return fh, err
}

// MultipartForm is the parsed multipart form, including file uploads.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMaxMultipartSize)
	if c.IsFrozenRequestBody {
		c.ResetBodyReaderPosition()
	}
	return c.Request.MultipartForm, err
}

// Cookie returns the named cookie provided in the request or ErrNoCookie if not found.
// And return the named cookie is unescaped.
// If multiple cookies match the given name, only one cookie will be returned.
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

// ClientIP implements a best effort algorithm to return the real client IP, it parses X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
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

// Get user agent from request headers.
func (c *Context) UserAgent() string {
	return c.MustRequestHeader("User-Agent")
}

// Get the Content-Type header of the request.
func (c *Context) ContentType() string {
	contentType := c.MustRequestHeader(ContentTypeHeaderKey)
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

// Get url scheme from request, supported X-Scheme and X-Forwarded-Proto headers.
func (c *Context) RequestUrlScheme() string {
	if value, ok := c.RequestHeader("X-Scheme"); ok && len(value) > 0 {
		return strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := c.RequestHeader("X-Forwarded-Proto"); ok && len(value) > 0 && value != "http" {
		return strings.ToLower(strings.TrimSpace(value))
	}
	if c.Request.URL != nil {
		if len(c.Request.URL.Scheme) > 1 {
			return c.Request.URL.Scheme
		}
	}
	return "http"
}

// Get request header string value by key.
func (c *Context) RequestHeader(key string) (string, bool) {
	if values, ok := c.Request.Header[key]; len(values) > 0 && ok {
		return values[0], true
	}
	return "", false
}

// Get request header string value by key with default value.
func (c *Context) RequestHeaderDef(key string, def string) string {
	if value, ok := c.RequestHeader(key); ok {
		return value
	}
	return def
}

func (c *Context) MustRequestHeader(key string) string {
	return c.RequestHeaderDef(key, "")
}

// Reset body reader position.
func (c *Context) ResetBodyReaderPosition() error {
	_, err := topSeekReader(c.Request.Body, true)
	return err
}

func (c *Context) IsLocalRequest() bool {
	return c.isLocalRequest
}
