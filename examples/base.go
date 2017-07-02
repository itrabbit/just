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
		POST("/{name}", nil, func(context *just.Context) just.IResponse {
			postValue := context.PostFormDef("value", "")
			return &just.Response{201, []byte("Hi " + context.GetParamDef("name", "unknown") + ", post value = " + postValue), nil}
		})
	app.Run("127.0.0.1:8000")
}
