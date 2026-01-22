package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range dirEntries {
		// Если файл является директорией, то продолжаем (переходим к следующему файлу в цикле)
		if file.IsDir() {
			continue
		}

		// Чтобы каждый раз не обращаться к Name(), присвоим имя переменной fileName
		fileName := file.Name()
		// Если попался файл, в имени которого есть знак '=', возвращаем соответствующую ошибку
		if strings.Contains(fileName, "=") {
			return nil, fmt.Errorf("env file name contains '=' character: %s", fileName)
		}

		// Переходим к считыванию значения переменной из файла
		// Для удобства присвоим переменной filePath путь к файлу в ОС
		filePath := filepath.Join(dir, fileName)
		osFile, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		// По окончании работы с открытым файлом закрыть его
		defer osFile.Close()
		// Создаем reader и считываем первую строку
		reader := bufio.NewReader(osFile)
		fString, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("eror by reading of file: %w", err)
		}
		fString = strings.TrimSuffix(fString, "\n")
		// Удаляем пробелы и табуляции справа
		fString = strings.TrimRight(fString, "\t ")
		// Терминальные нули (0x00) заменяются на перевод строки (\n)
		fString = string(bytes.ReplaceAll([]byte(fString), []byte{0x00}, []byte("\n")))

		// Записываем результат в мапу env
		if len(fString) == 0 && errors.Is(err, io.EOF) {
			env[fileName] = EnvValue{NeedRemove: true}
		} else {
			env[fileName] = EnvValue{Value: fString, NeedRemove: false}
		}
	}
	return env, nil
}
