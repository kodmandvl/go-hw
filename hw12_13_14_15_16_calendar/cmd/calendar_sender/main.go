package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/spf13/pflag"
)

var configFile string

func init() {
	pflag.StringVarP(&configFile, "config", "c", "./configs/sender_config.yaml", "Path to configuration file")
}

func main() {
	pflag.Parse()

	if pflag.Arg(0) == "version" {
		printVersion()
		return
	}

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	cfg := NewConfig()
	logg := logger.New(cfg.Logger.Level, os.Stdout)

	cons, err := rabbitmq.NewConsumer(cfg.Rabbit.URI, cfg.Rabbit.Queue)
	if err != nil {
		logg.Error("connect to RabbitMQ: %s", err.Error())
		return fmt.Errorf("rabbitmq consumer: %w", err)
	}
	defer cons.Close()

	// Publisher для статусов отправки (не обязателен): позволяет интеграционным тестам
	// проверить факт обработки уведомления sender-ом через отдельную очередь.
	var statusPub rabbitmq.Publisher
	if cfg.Rabbit.StatusQueue != "" {
		statusPub, err = rabbitmq.NewPublisher(cfg.Rabbit.URI, cfg.Rabbit.StatusQueue)
		if err != nil {
			logg.Error("connect to RabbitMQ status queue: %s", err.Error())
			return fmt.Errorf("rabbitmq status publisher: %w", err)
		}
		defer statusPub.Close()
	}

	logg.Info("calendar_sender is running, waiting for messages...")

	handler := func(ctx context.Context, body []byte) error {
		// По заданию достаточно вывести полезную нагрузку в лог (как "имитация" отправки).
		logg.Info("notification payload: %s", string(body))

		if statusPub != nil {
			status := map[string]string{
				"status":         "sent",
				"processed_at":   time.Now().UTC().Format(time.RFC3339),
				"notification":   string(body),
				"sender_service": "calendar_sender",
			}
			statusBody, marshalErr := json.Marshal(status)
			if marshalErr != nil {
				return fmt.Errorf("marshal status payload: %w", marshalErr)
			}
			if publishErr := statusPub.Publish(ctx, statusBody); publishErr != nil {
				return fmt.Errorf("publish status payload: %w", publishErr)
			}
		}

		return nil
	}

	if err := cons.Consume(ctx, handler); err != nil && ctx.Err() == nil {
		return err
	}

	return nil
}
