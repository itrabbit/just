package finalizer

import (
	"reflect"
	"strings"
	"time"

	"github.com/itrabbit/just"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

const (
	// Типы условий исключения из выдачи
	excludeCmEqual = 1
)

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		if m := v.MethodByName("IsZero"); m.IsValid() {
			return m.Call([]reflect.Value{})[0].Bool()
		}
	}
	return false
}

func checkExportedInterface(value reflect.Value) bool {
	kind := value.Kind()
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func indexOfStrings(list []string, x string) int {
	if list != nil {
		for i, item := range list {
			if strings.EqualFold(item, x) {
				return i
			}
		}
	}
	return -1
}

type fieldOptions struct {
	Name              string
	Group             string
	Export            string
	Omitempty         bool
	Skip              bool
	ExcludeConditions map[string]int
}

func (o *fieldOptions) SetExcludeConditions(toFieldName string, opID int) {
	if o.ExcludeConditions == nil {
		o.ExcludeConditions = make(map[string]int)
	}
	o.ExcludeConditions[toFieldName] = opID
}

func parseFieldTags(encTagName string, field *reflect.StructField, groups ...string) *fieldOptions {
	options := &fieldOptions{}
	// 1. Определяемся с параметром пропуска значения
	groupTag := strings.TrimSpace(field.Tag.Get("group"))
	if len(groupTag) > 0 {
		if items := strings.Split(groupTag, ","); len(items) > 0 {
			options.Skip = true
			if len(groups) > 0 {
				for _, item := range items {
					if i := indexOfStrings(groups, item); i >= 0 {
						options.Skip = false
						break
					}
				}
			}
		}
	}
	if options.Skip {
		return options
	}
	// 2. Находим имя и определяем возможность пропуска пустого значения
	options.Name = field.Name
	if items := strings.Split(field.Tag.Get(encTagName), ","); len(items) > 0 {
		first := strings.TrimSpace(items[0])
		if len(first) > 0 {
			options.Name = first
			if options.Name == "-" {
				options.Skip = true
				return options
			}
		}
		if len(items) > 1 {
			for i, item := range items {
				if i > 0 {
					if strings.TrimSpace(item) == "omitempty" {
						options.Omitempty = true
						break
					}
				}
			}
		}
	}
	// 3. Получаем поле для экспорта внутреннего значения, если оно указано
	options.Export = strings.TrimSpace(field.Tag.Get("export"))
	// 4. Получаем указания для группысвнутренней финализации
	options.Group = strings.TrimSpace(field.Tag.Get("bygroup"))
	// 5. Получаем данные для условий исключения
	for _, exclude := range strings.Split(strings.TrimSpace(field.Tag.Get("exclude")), ";") {
		params := strings.Split(exclude, ":")
		if len(params) > 1 {
			switch strings.ToLower(strings.TrimSpace(params[0])) {
			case "equal":
				options.SetExcludeConditions(strings.TrimSpace(params[1]), excludeCmEqual)
				break
			default:
				break
			}
		}
	}
	return options
}

// Finalization of objects for a particular data model output to the serializer.
func Finalize(encTagName string, v interface{}, groups ...string) interface{} {
	// TODO: Оптимизировать в будущем
	val := reflect.Indirect(reflect.ValueOf(v))
	kind := val.Type().Kind()
	if kind == reflect.Array || kind == reflect.Slice {
		list := make([]interface{}, val.Len(), val.Len())
		for i := 0; i < val.Len(); i++ {
			if elem := val.Index(i); elem.IsValid() {
				list[i] = Finalize(encTagName, elem.Interface(), groups...)
				continue
			}
			list[i] = nil
		}
		return list
	} else if kind == reflect.Struct {
		t := val.Type()
		if !t.AssignableTo(timeType) && t.Name() != "Time" {
			m := make(just.H)
			// Перебираем поля и проверяем по группам
			for i := 0; i < t.NumField(); i++ {
				if field := t.Field(i); !field.Anonymous {
					if options := parseFieldTags(encTagName, &field, groups...); !options.Skip {
						fieldVal := val.Field(i)
						if options.Omitempty && isEmptyValue(fieldVal) {
							continue
						}
						if options.ExcludeConditions != nil && len(options.ExcludeConditions) > 0 {
							excluded := false
							for toFieldName, opID := range options.ExcludeConditions {
								if toFieldType, ok := t.FieldByName(toFieldName); ok && !toFieldType.Anonymous {
									// Сравнение значений
									if opID == excludeCmEqual {
										if toFieldVal := val.FieldByName(toFieldName); toFieldVal.IsValid() {
											if reflect.DeepEqual(fieldVal.Interface(), toFieldVal.Interface()) {
												excluded = true
												break
											}
										}
									}
								}
							}
							if excluded {
								continue
							}
						}
						if len(options.Export) > 0 {
							if indFieldVal := reflect.Indirect(fieldVal); indFieldVal.IsValid() {
								expKind := indFieldVal.Type().Kind()
								if expKind == reflect.Struct {
									if subField, ok := indFieldVal.Type().FieldByName(options.Export); ok {
										subFieldVal := indFieldVal.FieldByName(options.Export)
										if options.Omitempty && isEmptyValue(subFieldVal) {
											continue
										}
										if !subField.Anonymous && checkExportedInterface(subFieldVal) {
											if len(options.Group) > 0 {
												m[options.Name] = Finalize(encTagName, subFieldVal.Interface(), options.Group)
											} else {
												m[options.Name] = Finalize(encTagName, subFieldVal.Interface(), groups...)
											}
										} else if !options.Omitempty {
											m[options.Name] = nil
										}
									}
								} else if expKind == reflect.Slice || expKind == reflect.Array {
									arr := make([]interface{}, 0)
									for index := 0; index < indFieldVal.Len(); index++ {
										if itemVal := reflect.Indirect(indFieldVal.Index(index)); itemVal.Type().Kind() == reflect.Struct {
											if subField, ok := itemVal.Type().FieldByName(options.Export); ok {
												subFieldVal := itemVal.FieldByName(options.Export)
												if options.Omitempty && isEmptyValue(subFieldVal) {
													continue
												}
												if !subField.Anonymous && checkExportedInterface(subFieldVal) {
													if len(options.Group) > 0 {
														arr = append(arr, Finalize(encTagName, subFieldVal.Interface(), options.Group))
													} else {
														arr = append(arr, Finalize(encTagName, subFieldVal.Interface(), groups...))
													}
												}
											}
										}
									}
									if len(arr) > 0 || options.Omitempty {
										m[options.Name] = arr
									}
								}
							}
							continue
						}
						if checkExportedInterface(fieldVal) {
							if len(options.Group) > 0 {
								m[options.Name] = Finalize(encTagName, fieldVal.Interface(), options.Group)
							} else {
								m[options.Name] = Finalize(encTagName, fieldVal.Interface(), groups...)
							}
						} else if !options.Omitempty {
							m[options.Name] = nil
						}
					}
				}
			}
			return m
		}
	} else if kind == reflect.Map {
		for _, key := range val.MapKeys() {
			if elemValue := reflect.Indirect(val.MapIndex(key)); elemValue.IsValid() {
				elemKind := elemValue.Type().Kind()
				if elemKind == reflect.Interface ||
					elemKind == reflect.Struct ||
					elemKind == reflect.Slice ||
					elemKind == reflect.Array ||
					elemKind == reflect.Map {
					ptr := elemValue.Interface()
					if res := Finalize(encTagName, ptr, groups...); res != nil && res != ptr {
						val.SetMapIndex(key, reflect.ValueOf(res))
					}
				}
			}
		}
		return val.Interface()
	}
	return v
}
