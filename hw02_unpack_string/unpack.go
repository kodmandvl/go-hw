package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

// Помимо самой функции Unpack, у нас будут еще функции, которые Unpack вызывает.
// Изначально сделал всё в функции Unpack, но линтер рекомендовал переделать и декомпозировать.
// Помимо strconv.Atoi, использовал еще unicode.IsDigit.

func Unpack(input string) (string, error) {
	var builder strings.Builder // Используем strings.Builder
	runes := []rune(input)      // Преобразуем строку в слайс рун
	var prevRune rune           // Предудыщая руна в РАСПАКОВЫВАЕМОЙ строке
	var prevEscape bool         // Флажок, был ли в ЗАПАКОВАННОЙ строке предыдущий символ escape-символом бэк-слэш

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Если предыдущий символ был escape-символом бэк-слэш.
		if prevEscape {
			// В таком случае сначала проверка, что после него сейчас идет бэк-слэш или цифра.
			if err := escapedRuneCheck(r); err != nil {
				return "", err
			}
			// То есть если у нас сейчас символ бэкслэша или цифра, добавляем в нашу строку.
			builder.WriteRune(r)
			prevRune = r
			prevEscape = false
			continue // идем дальше
		}

		// Если у нас сейчас escape-символ бэк-слэш (в следующей итерации будет prevEscape = true).
		if r == '\\' {
			prevEscape = true
			// При этом этот бэк-слэш у нас не будет добавляться в распакованную строку, поэтому prevRune не переприсваиваем.
			continue // идем дальше
		}

		// Если у нас сейчас цифра и предыдущий сивол не был escape-символом бэкслэша.
		if unicode.IsDigit(r) {
			// Если цифра идет сразу в начале строки, то ошибка.
			if prevRune == 0 {
				return "", ErrInvalidString
			}
			// Если за цифрой следует цифра (то есть получаем двузначное число), то ошибка.
			if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				return "", ErrInvalidString
			}
			// Обрабатываем нашу руну-цифру (и prevRune не переприсваиваем).
			if err := digitProcess(&builder, prevRune, r); err != nil {
				return "", err
			}
			continue // идем дальше
		}

		// В прочих случаях просто добавялем символ в нашу строку и переприсваиваем prevRune.
		builder.WriteRune(r)
		prevRune = r
	}

	// Если уже после выхода из цикла у нас prevEscape = true.
	// То есть если последний символ в запакованной строке был исключающий символ бэк-слэш.
	// То ошибка.
	if prevEscape {
		return "", ErrInvalidString
	}

	// Возвращаем полученную строку:
	return builder.String(), nil
}

// Функция для проверки, что у нас после бэк-слэша идет исключаемый символ.
// Это или бэкслэш, или цифра, иначе ошибка.
func escapedRuneCheck(r rune) error {
	if r != '\\' && !unicode.IsDigit(r) {
		return ErrInvalidString
	}
	return nil
}

// Функция для обработки руны-цифры.
func digitProcess(builder *strings.Builder, prevRune, digit rune) error {
	count, err := strconv.Atoi(string(digit))
	if err != nil {
		return ErrInvalidString
	}
	// Если наша цифра - ноль:
	if count == 0 {
		s := builder.String()
		runes := []rune(s)
		if len(runes) == 0 {
			return ErrInvalidString
		}
		// Переписываем распаковываемую строку через взятие слайса рун от нее, не включая последний символ (он 0 раз).
		builder.Reset()
		builder.WriteString(string(runes[:len(runes)-1])) // от начала до последнего символа, не включая его
	} else {
		// Добавляем в распаковываемую стркоу символ prevRune еще count-1 раз, используя strings.Repeat.
		builder.WriteString(strings.Repeat(string(prevRune), count-1))
	}
	return nil
}
