package main

import (
	"github.com/itrabbit/just"
)

type TestForm struct {
	Name string `form:"name"`
}

func main() {
	app := just.New()
	app.Group("/v1").
		GET("/{path:path}", func(c *just.Context) just.IResponse {
			return c.ResponseDataFast(201, just.H{
				"path": "/" + c.GetParamDef("path", ""),
			})
		}).
		GET("/v", func(c *just.Context) just.IResponse {
			return just.RedirectResponse(301, "http://google.com")
		})

	app.Run("127.0.0.1:8000")
}
