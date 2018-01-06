package just

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

/**
 * Note: the validator relies on regular expression patterns of routing (bool, int, float, uuid).
 */

var (
	rxValidUuid    = regexp.MustCompile(patternParamUUID)
	rxValidFloat   = regexp.MustCompile(patternParamFloat)
	rxValidBoolean = regexp.MustCompile(patternParamBoolean)
	rxValidInteger = regexp.MustCompile(patternParamInteger)
)

// Structure validation errors.
type ValidationError struct {
	Field   string
	Message string
}

// Validation error text.
func (e *ValidationError) Error() string {
	if len(e.Field) > 0 && len(e.Message) > 0 {
		return "Invalid \"" + e.Field + "\" - " + e.Message
	}
	return "Unknown"
}

func parseInstruction(instruction string) (string, string) {
	if start := strings.Index(instruction, "("); start > 0 {
		if name := strings.ToLower(strings.TrimSpace(instruction[:start])); len(name) > 0 {
			return name, strings.TrimSpace(strings.Trim(instruction[start:], "()"))
		}
	}
	return strings.ToLower(strings.TrimSpace(instruction)), ""
}

func validationInt(i int64, instruction string) error {
	if name, value := parseInstruction(instruction); len(name) > 0 && len(value) > 0 {
		if name[0] == 'm' {
			if m, err := strconv.ParseInt(value, 10, 64); err == nil {
				if name == "min" {
					if i < m {
						return errors.New("value < min value")
					}
				} else if i > m {
					return errors.New("value > max value")
				}
			}
		}
	}
	return nil
}

func validationUnsignedInt(i uint64, instruction string) error {
	if name, value := parseInstruction(instruction); len(name) > 0 && len(value) > 0 {
		if name[0] == 'm' {
			if m, err := strconv.ParseUint(value, 10, 64); err == nil {
				if name == "min" {
					if i < m {
						return errors.New("value < min value")
					}
				} else if i > m {
					return errors.New("value > max value")
				}
			}
		}
	}
	return nil
}

func validationFloat(f float64, instruction string) error {
	if name, value := parseInstruction(instruction); len(name) > 0 && len(value) > 0 {
		if name[0] == 'm' {
			if m, err := strconv.ParseFloat(value, 64); err == nil {
				if name == "min" {
					if f < m {
						return errors.New("value < min value")
					}
				} else if f > m {
					return errors.New("value > max value")
				}
			}
		}
	}
	return nil
}

func validationString(str string, instruction string) error {
	if name, value := parseInstruction(instruction); len(name) > 0 {
		switch name {
		case "boolean", "bool", "b":
			if !rxValidBoolean.MatchString(str) {
				return errors.New("is not boolean")
			}
		case "integer", "int", "i":
			if !rxValidInteger.MatchString(str) {
				return errors.New("is not integer")
			}
		case "float", "number", "f":
			if !rxValidFloat.MatchString(str) {
				return errors.New("is not float")
			}
		case "uuid":
			if !rxValidUuid.MatchString(str) {
				return errors.New("is not UUID")
			}
		case "rgx", "regexp":
			if rx, err := regexp.Compile(value); err == nil && rx != nil {
				if !rx.MatchString(str) {
					return errors.New("is not valid by regexp pattern " + value)
				}
			}
		}
	}
	return nil
}

// Validation of the structure of the object according to the parameters within the tags.
func Validation(obj interface{}) []error {
	var result []error = nil
	if obj != nil {
		if val := reflect.Indirect(reflect.ValueOf(obj)); val.IsValid() {
			t := val.Type()
			if k := t.Kind(); k != reflect.Array && k != reflect.Slice {
				// Перебираем поля модели
				for i := 0; i < t.NumField(); i++ {
					if field := t.Field(i); len(field.Tag) > 0 && !field.Anonymous {
						// Получаем инструкции валидации
						if str, ok := field.Tag.Lookup("valid"); ok && len(str) > 0 {
							for _, instruction := range strings.Split(str, ";") {
								var err error = nil
								// Убираем пробелы
								instruction = strings.TrimSpace(instruction)
								// Проводим анализ
								fieldKind := field.Type.Kind()
								switch fieldKind {
								// Для целых чисел (min,max)
								case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
									err = validationInt(val.Field(i).Int(), instruction)
								case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
									err = validationUnsignedInt(val.Field(i).Uint(), instruction)
								case reflect.Float32, reflect.Float64:
									err = validationFloat(val.Field(i).Float(), instruction)
								case reflect.String:
									err = validationString(val.Field(i).String(), instruction)
								default:
									err = errors.New("Unsupported validator")
								}
								if err != nil {
									if result == nil {
										result = make([]error, 0)
									}
									result = append(result, &ValidationError{
										Field:   field.Name,
										Message: err.Error(),
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return result
}
