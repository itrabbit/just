package main

import (
	"github.com/itrabbit/just"
)

func main() {
	app := just.New()
	app.Group("/v1").
		GET("/{name}/{id:regexp(\\d+)}", func(c *just.Context) just.IResponse {
			return c.ResponseData(
				"xml", 201, just.H{
					"name": c.GetParamDef("name", ""),
					"id":   c.GetIntParamDef("id", 0),
				})
		})
	app.Run("127.0.0.1:8000")
}
