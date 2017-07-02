package main

import (
	"github.com/itrabbit/just"
)

func main() {
	app := just.New()
	app.Group("/v1").
		Use(func(context *just.Context) just.IResponse {
			return context.Next()
		}).
		GET("/{name}/{id:regexp(\\d+)}", func(context *just.Context) just.IResponse {
			return &just.Response{201, []byte("Hi " + context.GetParamDef("name", "unknown") + ", id = " + context.GetParamDef("id", "0")), nil}
		})
	app.Run("127.0.0.1:8000")
}
