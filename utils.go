package just

import (
	"path"
)

type (
	H map[string]interface{}
)

// joinPaths - сложение путей
func joinPaths(a string, b string) string {
	if len(b) < 1 {
		return a
	}
	c := path.Join(a, b)
	if b[len(b)-1] == '/' && c[len(c)-1] != '/' {
		return c + "/"
	}
	return c
}
