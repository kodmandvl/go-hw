package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 0
	}

	// Создаем мапу переменных среды и забираем из среды текущие заданные переменные,
	// кроме тех, что заданы как пустая строка, например "MY_ENV_VAL="
	envMap := map[string]string{}
	for _, kvEnv := range os.Environ() {
		kvEnvParts := strings.SplitN(kvEnv, "=", 2)
		if len(kvEnvParts) == 2 {
			envMap[kvEnvParts[0]] = kvEnvParts[1]
		}
	}

	// Далее сверяемся уже с мапой env (полученной из ReadDir)
	for k, v := range env {
		// Если NeedRemove - true, удаляем из мапы envMap
		// В противном случае добавляем в мапу envMap
		if v.NeedRemove {
			delete(envMap, k)
		} else {
			envMap[k] = v.Value
		}
	}

	// Далее уже с набором актуальных значений переменных создаем слайс строк "ENV_VAR=VALUE"
	actualEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		actualEnv = append(actualEnv, fmt.Sprintf("%s=%s", k, v))
	}

	// Формирование команды и окружения и запуск
	// #nosec G204
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Env = actualEnv
	if err := execCmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// fmt.Println("ERROR MATCH")
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
