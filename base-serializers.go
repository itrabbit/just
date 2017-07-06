package just

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"net/url"
	"regexp"
)

/** JSON Serializer */

type JsonSerializer struct{}

func (JsonSerializer) DefaultContentType() string {
	return "application/json"
}

func (JsonSerializer) Serialize(v interface{}) ([]byte, error) {
	if IsDebug() {
		return json.MarshalIndent(v, "", "    ")
	}
	return json.Marshal(v)
}

func (JsonSerializer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

/** XML Serializer */

type XmlSerializer struct{}

func (XmlSerializer) DefaultContentType() string {
	return "application/xml"
}

func (XmlSerializer) Serialize(v interface{}) ([]byte, error) {
	if IsDebug() {
		return xml.MarshalIndent(v, "", "    ")
	}
	return xml.Marshal(v)
}

func (XmlSerializer) Deserialize(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

/** Form Serializer */

const (
	defaultMaxMultipartSize = 32 << 20
	defaultMaxUrlSize       = 10 << 20
)

var (
	urlencodedRx = regexp.MustCompile("([A-Za-z0-9%./]+=[^\\s]+)")
)

type FormSerializer struct{}

func (FormSerializer) DefaultContentType() string {
	return "application/x-www-form-urlencoded"
}

func (FormSerializer) Serialize(v interface{}) ([]byte, error) {
	return marshalUrlValues(v)
}

func (FormSerializer) Deserialize(data []byte, v interface{}) error {
	if len(data) <= defaultMaxUrlSize && urlencodedRx.Match(data) {
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}
		return mapForm(values, nil, v)
	}
	if end := bytes.LastIndex(data, []byte("--")); end > 0 {
		if start := bytes.LastIndex(data[:end], []byte("\n--")); start > 0 && end > start {
			if boundary := string(data[start:end]); len(boundary) > 0 {
				r := multipart.NewReader(bytes.NewReader(data), boundary)
				form, err := r.ReadForm(defaultMaxMultipartSize)
				if err != nil {
					return err
				}
				return mapForm(form.Value, form.File, v)
			}
		}
	}
	return nil
}
