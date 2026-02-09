package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

type UserEmail struct {
	Email string `json:"email"`
}

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	resMap := make(DomainStat)
	scanner := bufio.NewScanner(r)
	suffix := "." + domain

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // Пропускаем пустую строку и переходим к следующей
		}

		var uEmail UserEmail
		if err := json.Unmarshal(line, &uEmail); err != nil {
			continue // Пропускаем некорректную строку и переходим к следующей
		}

		email := uEmail.Email
		if !strings.HasSuffix(email, suffix) {
			continue // Указанный суффикс не нашли - переходим к следующей строке
		}

		// Извлекаем домен
		ind := strings.LastIndex(email, "@")
		if ind == -1 {
			continue // Несмотря на то, что мы ранее нашли суффикс в строке, в ней нет "@" в имейле, переходим к следующей строке
		}

		// Добавляем доменное имя имейла в реузультирующую мапу и инкрементируем
		domainName := strings.ToLower(email[ind+1:])
		resMap[domainName]++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return resMap, nil
}
