package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/spf13/pflag"
)

var configFile string

func init() {
	// pflag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
	pflag.StringVarP(&configFile, "config", "c", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	// Использую pflag вместо flag.
	pflag.Parse()

	if pflag.Arg(0) == "version" {
		printVersion()
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := NewConfig()
	logg := logger.New(config.Logger.Level, os.Stdout)

	var eventStorage storage.EventStorage

	if config.Storage.Type == "sql" {
		logg.Info("sql storage selected")
		connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.DB.DBUsername, config.DB.DBPassword, config.DB.DBHost, config.DB.DBPort, config.DB.DBName)
		// logg.Info("connectionString: " + connectionString)
		eventStorage = sqlstorage.New(connectionString)
		err := eventStorage.Connect(ctx)
		if err != nil {
			logg.Error("connect to DBMS server: %s", err.Error())
			cancel()
			os.Exit(1) //nolint:gocritic
		}
		defer eventStorage.Close()
	} else {
		logg.Info("memory storage selected")
		eventStorage = memorystorage.New()
	}

	logg.Info("init %s storage: OK", config.Storage.Type)

	calendar := app.New(logg, eventStorage)

	server := internalhttp.NewServer(config.HTTPServer.Host, config.HTTPServer.Port, logg, calendar)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("stop http server: %s", err.Error())
		}
	}()

	logg.Info("calendar is running...")
	// logg.Error("test error message...")
	// logg.Warning("test warning message...")
	// logg.Debug("test debug message...")

	if err := server.Start(ctx); err != nil {
		logg.Error("start http server: %s", err.Error())
		cancel()
		os.Exit(1)
	}
}
