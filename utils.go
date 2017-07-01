package just

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
)

type (
	H map[string]interface{}
)

func topSeekReader(reader io.Reader, nopDetect bool) (int64, error) {
	if reader != nil {
		if nopDetect {
			if val := reflect.ValueOf(reader); val.IsValid() {
				if val.Type().Name() == "nopCloser" {
					if val = val.FieldByName("Reader"); val.IsValid() {
						return topSeekReader(val.Interface().(io.Reader), false)
					}
				}
			}
		}
		if b, ok := reader.(*bytes.Reader); ok {
			return b.Seek(0, io.SeekStart)
		} else if s, ok := reader.(*strings.Reader); ok {
			return s.Seek(0, io.SeekStart)
		} else if f, ok := reader.(*os.File); ok {
			return f.Seek(0, io.SeekStart)
		}
		return -1, errors.New("Reader is not supported")
	}
	return -1, errors.New("Reader is empty")
}

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
