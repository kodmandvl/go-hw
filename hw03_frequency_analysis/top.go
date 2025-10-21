package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

// Функции regexp.MustCompile и strings.Split не использовал.
// Но они понадобились бы при выполнении задания со (*).
func Top10(text string) []string {
	if len(text) == 0 {
		return nil
	}

	// Разбиваем текст на слова по пробелам, функция strings.Fields возвращает слайс из слов.
	words := strings.Fields(text)

	// Считаем частоты с помощью Map со словами в качестве ключей и частотами в качестве значений.
	frequencyMap := make(map[string]int)
	// Проходим слайс слов words и инкрементируем частоты frequencyMap (индексы слайса не важны, важны только значения).
	for _, word := range words {
		frequencyMap[word]++
	}

	// Создаем слайс для сортировки.
	// Это будет слайс структур.
	type wordCount struct {
		word  string
		count int
	}
	wc := make([]wordCount, 0, len(frequencyMap))
	for w, c := range frequencyMap {
		wc = append(wc, wordCount{w, c})
	}

	// Теперь сортируем: сначала по убыванию count, потом лексикографически по слову.
	// В документации пример по использованию sort.Slice: https://pkg.go.dev/sort#Slice
	sort.Slice(wc, func(i, j int) bool {
		// Если одинаковое количество раз, то сортируем слова лексикографически:
		if wc[i].count == wc[j].count {
			return wc[i].word < wc[j].word
		}
		// А если же количество разное, то от большего к меньшему сортируем:
		return wc[i].count > wc[j].count
	})

	// Возвращаем топ10 слов по частоте (или меньше, если всего слов нашлось меньше 10).
	n := 10
	if len(wc) < 10 {
		n = len(wc)
	}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = wc[i].word
	}
	return result
}
