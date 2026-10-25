package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Validate(v any) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("нулевой указатель")
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ожидалась структура, получен %s", val.Kind())
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ";")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			parts := strings.SplitN(rule, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("некорректное правило %q у поля %s", rule, field.Name)
			}

			key := strings.TrimSpace(parts[0])
			raw := strings.TrimSpace(parts[1])

			switch key {
			case "min":
				min, err := strconv.Atoi(raw)
				if err != nil {
					return fmt.Errorf("некорректное значение min у поля %s: %w", field.Name, err)
				}
				if err := checkMin(field.Name, fieldVal, min); err != nil {
					return err
				}
			case "max":
				max, err := strconv.Atoi(raw)
				if err != nil {
					return fmt.Errorf("некорректное значение max у поля %s: %w", field.Name, err)
				}
				if err := checkMax(field.Name, fieldVal, max); err != nil {
					return err
				}
			case "regexp":
				if fieldVal.Kind() != reflect.String {
					return fmt.Errorf("regexp применим только к строкам, поле %s", field.Name)
				}
				re, err := regexp.Compile(raw)
				if err != nil {
					return fmt.Errorf("некорректный regexp у поля %s: %w", field.Name, err)
				}
				s := fieldVal.String()
				if !re.MatchString(s) {
					return fmt.Errorf("поле %s не соответствует regexp", field.Name)
				}
			default:
				return fmt.Errorf("неизвестное правило %q у поля %s", key, field.Name)
			}
		}
	}

	return nil
}

func checkMin(name string, fieldVal reflect.Value, min int) error {
	switch fieldVal.Kind() {
	case reflect.String:
		length := len([]rune(fieldVal.String()))
		if length < min {
			return fmt.Errorf("длина поля %s равна %d, меньше min %d", name, length, min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int(fieldVal.Int())
		if n < min {
			return fmt.Errorf("значение поля %s равно %d, меньше min %d", name, n, min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := int(fieldVal.Uint())
		if n < min {
			return fmt.Errorf("значение поля %s равно %d, меньше min %d", name, n, min)
		}
	default:
		return fmt.Errorf("min не поддерживается для поля %s типа %s", name, fieldVal.Kind())
	}
	return nil
}

func checkMax(name string, fieldVal reflect.Value, max int) error {
	switch fieldVal.Kind() {
	case reflect.String:
		length := len([]rune(fieldVal.String()))
		if length > max {
			return fmt.Errorf("длина поля %s равна %d, больше max %d", name, length, max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := int(fieldVal.Int())
		if n > max {
			return fmt.Errorf("значение поля %s равно %d, больше max %d", name, n, max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := int(fieldVal.Uint())
		if n > max {
			return fmt.Errorf("значение поля %s равно %d, больше max %d", name, n, max)
		}
	default:
		return fmt.Errorf("max не поддерживается для поля %s типа %s", name, fieldVal.Kind())
	}
	return nil
}

type User struct {
	Name  string `validate:"min=3"`
	Age   int    `validate:"min=18;max=65"`
	Email string `validate:"regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
}

func main() {
	if err := Validate(User{Name: "Ив", Age: 18, Email: "test@example.com"}); err != nil {
		fmt.Println("Ошибка валидации:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 70, Email: "test@example.com"}); err != nil {
		fmt.Println("Ошибка валидации:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 35, Email: "invalid email"}); err != nil {
		fmt.Println("Ошибка валидации:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 35, Email: "test@example.com"}); err != nil {
		fmt.Println("Ошибка валидации:", err)
	} else {
		fmt.Println("Валидация успешна")
	}
}
