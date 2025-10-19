// Просто main.go для запуска ручной проверки и отладки до этапа тестов
// Вот так вызываю из директории с модулем:
// go run main/main.go

package main

import (
	"bufio"
	"fmt"
	"os"

	hw02unpackstring "github.com/kodmandvl/go-hw/hw02_unpack_string" // ИМПОРТ
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Введите строку: ")
		if !scanner.Scan() {
			break // Прерываем, если ввод завершен (например, Ctrl+D)
		}
		input := scanner.Text()
		fmt.Println("Будет распаковка строки: ", input)
		s, e := hw02unpackstring.Unpack(input) // ВЫЗОВ
		if e == nil {
			fmt.Println(s)
		} else {
			fmt.Println("Error:", e)
		}
	}
}
