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
		GET("/{id:regexp(\\d+)}", func(c *just.Context) just.IResponse {
			var form TestForm
			c.Bind(&form)

			return c.ResponseDataFast(201, just.H{
				"id":   c.GetIntParamDef("id", 0),
				"name": form.Name,
			})
		}).
		GET("/v", func(c *just.Context) just.IResponse {
			return just.RedirectResponse(301, "http://google.com")
		})

	app.Run("127.0.0.1:8000")
}
