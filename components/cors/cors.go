package cors

import (
	"strconv"
	"strings"
	"time"

	"github.com/itrabbit/just"
)

var (
	defaultAllowHeaders = []string{
		"Origin", "Accept", "Content-Type",
		"Authorization", "X-Auth", "X-Scheme",
		"X-Real-Ip", "X-Forwarded-For", "X-Forwarded-Proto",
		"Token", "Connection", "Upgrade", "Sec-Websocket-Version",
		"Sec-Websocket-Key", "Sec-Websocket-Protocol", "Sec-Websocket-Accept",
	}
	defaultAllowMethods = []string{
		"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS",
	}
)

// CORS Options struct.
type Options struct {
	AllowOrigins     []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	MaxAge           time.Duration
}

// CORS middleware.
func Middleware(options Options) just.HandlerFunc {
	if options.AllowHeaders == nil {
		options.AllowHeaders = defaultAllowHeaders
	}
	if options.AllowMethods == nil {
		options.AllowMethods = defaultAllowMethods
	}
	return func(c *just.Context) just.IResponse {
		var (
			origin         = c.MustRequestHeader("Origin")
			requestMethod  = c.MustRequestHeader("Access-Control-Request-Method")
			requestHeaders = c.MustRequestHeader("Access-Control-Request-Headers")
		)
		if c.Request != nil && c.Request.Method == "OPTIONS" {
			// Блокируем дальнейшее выполнение и отправляем пустой ответ 200
			headers := make(map[string]string)
			if len(options.AllowMethods) > 0 {
				headers["Access-Control-Allow-Methods"] = strings.Join(options.AllowMethods, ",")
			} else if requestMethod != "" {
				headers["Access-Control-Allow-Methods"] = requestMethod
			}
			if len(options.AllowHeaders) > 0 {
				headers["Access-Control-Allow-Headers"] = strings.Join(options.AllowHeaders, ",")
			} else if requestHeaders != "" {
				headers["Access-Control-Allow-Headers"] = requestHeaders
			}
			if options.MaxAge > time.Duration(0) {
				headers["Access-Control-Max-Age"] = strconv.FormatInt(int64(options.MaxAge/time.Second), 10)
			}
			if len(options.AllowOrigins) > 0 {
				headers["Access-Control-Allow-Origin"] = strings.Join(options.AllowOrigins, " ")
			} else {
				headers["Access-Control-Allow-Origin"] = origin
			}
			if options.AllowCredentials {
				headers["Access-Control-Allow-Credentials"] = "true"
			}
			if len(options.ExposeHeaders) > 0 {
				headers["Access-Control-Expose-Headers"] = strings.Join(options.ExposeHeaders, ",")
			}
			return &just.Response{
				Status:  200,
				Bytes:   nil,
				Headers: headers,
			}
		}
		// Если не OPTION, вызваем следующие handlers
		res := c.Next()
		if res != nil {
			// Производим модификацию заголовков ответа
			if headers := res.GetHeaders(); headers != nil {
				if len(options.AllowOrigins) > 0 {
					headers["Access-Control-Allow-Origin"] = strings.Join(options.AllowOrigins, " ")
				} else {
					headers["Access-Control-Allow-Origin"] = origin
				}
				if options.AllowCredentials {
					headers["Access-Control-Allow-Credentials"] = "true"
				}
				if len(options.ExposeHeaders) > 0 {
					headers["Access-Control-Expose-Headers"] = strings.Join(options.ExposeHeaders, ",")
				}
			}
		}
		return res
	}
}
