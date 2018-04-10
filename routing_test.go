package just

import (
	"testing"
)

func TestRouterParamsInGroup(t *testing.T) {
	r := new(Router)
	group := r.Group("/{id:integer}", func(context *Context) IResponse {
		return context.Next()
	})
	route := group.GET("", func(context *Context) IResponse {
		return &Response{}
	})
	_, ok := route.CheckPath("/12")
	if !ok {
		t.Fail()
	}

}
