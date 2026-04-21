package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/spf13/pflag"
)

var configFile string

func init() {
	pflag.StringVarP(&configFile, "config", "c", "./configs/scheduler_config.yaml", "Path to configuration file")
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

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.DBUsername, cfg.DB.DBPassword, cfg.DB.DBHost, cfg.DB.DBPort, cfg.DB.DBName)

	store := sqlstorage.New(connectionString)
	if err := store.Connect(ctx); err != nil {
		logg.Error("connect to DBMS: %s", err.Error())
		return fmt.Errorf("connect to DBMS: %w", err)
	}
	defer store.Close()

	pub, err := rabbitmq.NewPublisher(cfg.Rabbit.URI, cfg.Rabbit.Queue)
	if err != nil {
		logg.Error("connect to RabbitMQ: %s", err.Error())
		return fmt.Errorf("rabbitmq publisher: %w", err)
	}
	defer pub.Close()

	interval := time.Duration(cfg.Scheduler.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}

	logg.Info("calendar_scheduler is running (interval=%s)...", interval.String())

	// Блокируемся до сигнала; внутри — тикер и обработка уведомлений/удалений.
	if err := scheduler.Run(ctx, logg, store, pub, interval); err != nil && ctx.Err() == nil {
		return err
	}

	return nil
}
