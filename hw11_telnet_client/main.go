package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	// Читаем таймаут подключения из флага.
	timeout := pflag.Duration("timeout", 10*time.Second, "timeout for connect")
	pflag.Parse()
	// fmt.Println("timeout: " + timeout.String())

	// Ожидаем аргументы host и port.
	if pflag.NArg() < 2 {
		fmt.Println("Usage: go-telnet [--timeout=<your_desired_timeout>] <host> <port>")
		os.Exit(1)
	}

	host := pflag.Arg(0)
	port := pflag.Arg(1)
	address := host + ":" + port
	// fmt.Println("address: " + address)

	// Контекст завершится по Ctrl+C или SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Источник ввода/вывода привязываем к stdin/stdout.
	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Err connection: %s\n", err)
		return
	}
	defer client.Close()

	// WaitGroup для 2 горутин.
	var wg sync.WaitGroup
	wg.Add(2)

	// Получаем данные от сервера и выводим в stdout.
	go func() {
		defer wg.Done()
		if err := client.Receive(); err != nil {
			fmt.Fprintln(os.Stderr, "...Connection closed by peer")
		}
	}()

	// Отправляем наш ввод на сервер.
	go func() {
		defer wg.Done()
		if err := client.Send(); err != nil {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	}()

	// По сигналу завершаем соединение и выходим.
	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "...Connection closed")
		client.Close()
	}()

	wg.Wait()
}
