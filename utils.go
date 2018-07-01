package just

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
)

// Errors
var (
	ErrEmptyReader       = errors.New("reader is empty")
	ErrUnsupportedReader = errors.New("reader is not supported")
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
		return -1, ErrUnsupportedReader
	}
	return -1, ErrEmptyReader
}

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

// H is a shortcup for map[string]interface{} (to improve code reading).
type H map[string]interface{}

// Overriding for Marshal to XML
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "data",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return nil
}
