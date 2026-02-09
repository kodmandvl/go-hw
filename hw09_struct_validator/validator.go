package hw09structvalidator

// Про рефлексию и примеры валидации структур есть также статья на Habr:
// https://habr.com/ru/companies/otus/articles/833770/

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Программные ошибки (неверный тэг, регулярка и пр.) и ошибки валидации.
var (
	// Программные ошибки.
	ErrNotStruct     = errors.New("interface is not struct")
	ErrInvalidTag    = errors.New("invalid validate tag")
	ErrInvalidRegexp = errors.New("invalid regexp")
	// Ошибки валидации.
	ErrLen    = errors.New("invalid length")
	ErrMin    = errors.New("too small")
	ErrMax    = errors.New("too large")
	ErrIn     = errors.New("not in allowed set")
	ErrRegexp = errors.New("regexp mismatch")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, e := range v {
		sb.WriteString(fmt.Sprintf("%s: %s\n", e.Field, e.Err.Error()))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	// Проверяем, что входной параметр является структурой.
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var errs ValidationErrors
	t := val.Type()

	// Проходим циклом по всем полям структуры
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		// Получаем тег validate. А если его нет, то пропускаем это поле и далее переходим к следующему.
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		fv := val.Field(i)
		// Пропускаем неэкспортируемые (приватные) поля (они не могут быть использованы через Interface()).
		if !fv.CanInterface() {
			continue
		}

		// Разделяем правила валидации по логическому "И".
		rules := strings.Split(tag, "|")
		for _, rule := range rules {
			// Парсим правило: "имя:аргумент".
			parts := strings.SplitN(rule, ":", 2)
			name := parts[0]
			arg := ""
			if len(parts) == 2 {
				arg = parts[1]
			}

			// Применяем правило к полю.
			if err := applyRule(fv, field.Name, name, arg); err != nil {
				var vErrs ValidationErrors
				// Если ошибка - это список ошибок валидации (ValidationErrors), то добавляем ошибки в общий список.
				if errors.As(err, &vErrs) {
					errs = append(errs, vErrs...)
				} else {
					// Иначе проверяем, является ли ошибка программной (ErrInvalidTag или ErrInvalidRegexp)
					if errors.Is(err, ErrInvalidTag) || errors.Is(err, ErrInvalidRegexp) {
						return err // Возвращаем программную ошибку напрямую
					}
					// Иначе создаем ValidationError для текущего поля.
					errs = append(errs, ValidationError{
						Field: field.Name,
						Err:   err,
					})
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// applyRuleприменяет правило валидации к значению поля.
func applyRule(fv reflect.Value, fieldName, name, arg string) error {
	// В зависимости от типа поля, вызываем соответствующую функцию валидации.
	switch fv.Kind() {
	case reflect.String:
		return validateString(fv.String(), name, arg)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateInt(int(fv.Int()), name, arg)
	case reflect.Slice:
		// Для слайсов валидируем каждый элемент отдельно.
		return validateSlice(fieldName, fv, name, arg)
	// Для остальных типов валидация не поддерживается - игнорируем.
	case reflect.Bool,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer, reflect.UnsafePointer,
		reflect.Struct,
		reflect.Invalid:
		return nil
	}
	return nil
}

// validateString применяет правила валидации к строковому значению.
func validateString(val, name, arg string) error {
	switch name {
	// Проверка длины строки.
	case "len":
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidTag, arg)
		}
		if len(val) != n {
			return fmt.Errorf("%w: expected %d, actual %d", ErrLen, n, len(val))
		}
	// Проверка соответствия регулярному выражению.
	case "regexp":
		re, err := regexp.Compile(arg)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidRegexp, arg)
		}
		if !re.MatchString(val) {
			return fmt.Errorf("%w: %s", ErrRegexp, arg)
		}
	// Проверка вхождения строки в допустимое множество.
	case "in":
		opts := strings.Split(arg, ",")
		for _, opt := range opts {
			if val == opt {
				return nil
			}
		}
		return fmt.Errorf("%w: must be one of %v", ErrIn, opts)
	}
	return nil
}

// validateInt применяет правила валидации к целочисленному значению.
func validateInt(val int, name, arg string) error {
	switch name {
	// Проверка минимального значения.
	case "min":
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("%w : %s", ErrInvalidTag, arg)
		}
		if val < n {
			return fmt.Errorf("%w: must be greater than or equal to %d", ErrMin, n)
		}
	// Проверка максимального значения.
	case "max":
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidTag, arg)
		}
		if val > n {
			return fmt.Errorf("%w: must be less than or equal to %d", ErrMax, n)
		}
	// Проверка вхождения числа в допустимое множество.
	case "in":
		opts := strings.Split(arg, ",")
		for _, opt := range opts {
			n, err := strconv.Atoi(opt)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidTag, opt)
			}
			if val == n {
				return nil
			}
		}
		return fmt.Errorf("%w: must be one of %v", ErrIn, opts)
	}
	return nil
}

// validateSlice применяет правило валидации к каждому элементу слайса.
// Возвращает ValidationErrors со всеми ошибками элементов.
func validateSlice(fieldName string, fv reflect.Value, name, arg string) error {
	var errs ValidationErrors
	// Выполняем applyRule для каждого элемента слайса.
	for i := 0; i < fv.Len(); i++ {
		elem := fv.Index(i)
		if err := applyRule(elem, fmt.Sprintf("%s[%d]", fieldName, i), name, arg); err != nil {
			errs = append(errs, ValidationError{
				Field: fmt.Sprintf("%s[%d]", fieldName, i),
				Err:   err,
			})
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
