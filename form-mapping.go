package just

import (
	"errors"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

func marshalUrlValues(ptr interface{}) ([]byte, error) {
	t, v := reflect.TypeOf(ptr).Elem(), reflect.ValueOf(ptr).Elem()
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return nil, errors.New("Array and slice by root element not supported, only structure!")
	}
	values := make(url.Values)
	for i := 0; i < t.NumField(); i++ {
		typeField, structField := t.Field(i), v.Field(i)
		if !structField.CanSet() {
			continue
		}
		structFieldKind, inputFieldName := structField.Kind(), typeField.Tag.Get("form")
		if len(inputFieldName) < 1 {
			inputFieldName = typeField.Name
		}
		if structFieldKind == reflect.Slice {
			for i := 0; i < structField.Len(); i++ {
				values.Add(inputFieldName, structField.Index(i).String())
			}
		} else {
			values.Set(inputFieldName, structField.String())
		}
	}
	return []byte(values.Encode()), nil
}

// mapForm - form mapping (+ multipart files support)
func mapForm(values map[string][]string, files map[string][]*multipart.FileHeader, ptr interface{}) error {
	return recursiveTreeMapForm(values, files, reflect.TypeOf(ptr).Elem(), reflect.ValueOf(ptr).Elem())
}

func recursiveTreeMapForm(values map[string][]string, files map[string][]*multipart.FileHeader, t reflect.Type, v reflect.Value) error {
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		return errors.New("Array and slice by root element not supported, only structure!")
	}
	for i := 0; i < t.NumField(); i++ {
		typeField, structField := t.Field(i), v.Field(i)
		if !structField.CanSet() {
			if typeField.Type.Kind() == reflect.Struct {
				if err := recursiveTreeMapForm(values, files, typeField.Type, structField); err != nil {
					return err
				}
			}
			continue
		}
		structFieldKind, inputFieldName := structField.Kind(), typeField.Tag.Get("form")
		if len(inputFieldName) < 1 {
			inputFieldName = typeField.Name
		}
		inputValue, existsValue := values[inputFieldName]
		if !existsValue {
			// Check files
			if files != nil && len(files) > 0 {
				if list, ok := files[inputFieldName]; ok {
					if numFiles := len(list); numFiles > 0 {
						if structFieldKind == reflect.Slice {
							if structField.Type().Elem().Name() == "FileHeader" {
								slice := reflect.MakeSlice(structField.Type(), numFiles, numFiles)
								for i := 0; i < numFiles; i++ {
									slice.Index(i).Set(reflect.ValueOf(list[i]))
								}
								v.Field(i).Set(slice)
							}
						} else if structField.Type().Elem().Name() == "FileHeader" {
							structField.Set(reflect.ValueOf(list[0]))
						}
					}
				}
			}
			continue
		}
		if numElements := len(inputValue); numElements > 0 {
			if structFieldKind == reflect.Slice || structFieldKind == reflect.Array {
				sliceOf := structField.Type().Elem().Kind()
				slice := reflect.MakeSlice(structField.Type(), numElements, numElements)
				for i := 0; i < numElements; i++ {
					if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
						return err
					}
				}
				v.Field(i).Set(slice)
			} else {
				if _, isTime := structField.Interface().(time.Time); isTime {
					if err := setTimeField(inputValue[0], typeField, structField); err != nil {
						return err
					}
					continue
				}
				kind := typeField.Type.Kind()
				if kind == reflect.Ptr {
					kind = typeField.Type.Elem().Kind()
					if structField.IsNil() {
						structField.Set(reflect.New(typeField.Type.Elem()))
					}
				}
				if err := setWithProperType(kind, inputValue[0], reflect.Indirect(structField)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return nil
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		return errors.New("Blank time format")
	}
	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}
	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}
	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(t))
	return nil
}
