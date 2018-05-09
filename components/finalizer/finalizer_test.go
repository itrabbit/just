package finalizer

import (
	"testing"

	"fmt"

	"github.com/itrabbit/just"
)

type ObjA struct {
	Name string `json:"name"`
}

type ObjM struct {
	Data just.H `json:"data,omitempty"`
}

func TestFinalize(t *testing.T) {
	obj := ObjM{
		Data: just.H{
			"obj": ObjA{
				Name: "test",
			},
		},
	}
	m := Finalize("json", &obj, "public", "private")
	fmt.Println(m)
}
