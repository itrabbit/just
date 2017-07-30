package main

import (
	"github.com/gin-gonic/gin"
	"github.com/itrabbit/just"
	"os"
	"github.com/itrabbit/just/components/cors"
	"strings"
	"time"
	"strconv"
	"net/http"
)

var (
	defaultAllowHeaders = []string{"Origin", "Accept", "Content-Type", "Authorization", "X-Auth", "Token"}
	defaultAllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
)

type CorsOptions struct {
	AllowOrigins     []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	MaxAge           time.Duration
}

func CorsMiddleware(options CorsOptions) gin.HandlerFunc {
	if options.AllowHeaders == nil {
		options.AllowHeaders = defaultAllowHeaders
	}
	if options.AllowMethods == nil {
		options.AllowMethods = defaultAllowMethods
	}
	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		origin := req.Header.Get("Origin")
		requestMethod := req.Header.Get("Access-Control-Request-Method")
		requestHeaders := req.Header.Get("Access-Control-Request-Headers")

		if len(options.AllowOrigins) > 0 {
			res.Header().Set("Access-Control-Allow-Origin", strings.Join(options.AllowOrigins, " "))
		} else {
			res.Header().Set("Access-Control-Allow-Origin", origin)
		}
		if options.AllowCredentials {
			res.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if len(options.ExposeHeaders) > 0 {
			res.Header().Set("Access-Control-Expose-Headers", strings.Join(options.ExposeHeaders, ","))
		}
		if req.Method == "OPTIONS" {
			if len(options.AllowMethods) > 0 {
				res.Header().Set("Access-Control-Allow-Methods", strings.Join(options.AllowMethods, ","))
			} else if requestMethod != "" {
				res.Header().Set("Access-Control-Allow-Methods", requestMethod)
			}
			if len(options.AllowHeaders) > 0 {
				res.Header().Set("Access-Control-Allow-Headers", strings.Join(options.AllowHeaders, ","))
			} else if requestHeaders != "" {
				res.Header().Set("Access-Control-Allow-Headers", requestHeaders)
			}
			if options.MaxAge > time.Duration(0) {
				res.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(options.MaxAge/time.Second), 10))
			}
			c.AbortWithStatus(http.StatusOK)
		} else {
			c.Next()
		}
	}
}

func main() {
	isGin := false
	for _, arg := range os.Args {
		if arg == "--gin" {
			isGin = true
			break
		}
	}
	if isGin {
		s0 := gin.Default()
		s0.Use(CorsMiddleware(CorsOptions{}))
		s0.GET("/welcome/:name", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"param_name": c.Param("name"),
				"framework":  "gin",
				"version":    gin.Version})
		})
		s0.Run(":8001")
	} else {
		s1 := just.New()
		s1.Use(cors.Middleware(cors.Options{}))
		s1.GET("/welcome/{name}", func(c *just.Context) just.IResponse {
			return c.Serializer().Response(200, just.H{
				"param_name": c.MustParam("name"),
				"framework":  "just",
				"version":    just.Version})
		})
		s1.Run(":8001")
	}
}
