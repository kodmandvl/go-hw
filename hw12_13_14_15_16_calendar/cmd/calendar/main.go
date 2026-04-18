package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/server/grpc"
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

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// run содержит основную логику, чтобы в main не вызывать os.Exit при активном defer (gocritic exitAfterDefer).
func run() error {
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
		if err := eventStorage.Connect(ctx); err != nil {
			logg.Error("connect to DBMS server: %s", err.Error())
			return fmt.Errorf("connect to DBMS: %w", err)
		}
		defer eventStorage.Close()
	} else {
		logg.Info("memory storage selected")
		eventStorage = memorystorage.New()
	}

	logg.Info("init %s storage: OK", config.Storage.Type)

	calendar := app.New(logg, eventStorage)

	// server := internalhttp.NewServer(config.HTTPServer.Host, config.HTTPServer.Port, logg, calendar)
	grpcAddr := fmt.Sprintf("%s:%d", config.GRPCServer.Host, config.GRPCServer.Port)
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logg.Error("grpc listen %s: %s", grpcAddr, err.Error())
		return fmt.Errorf("grpc listen: %w", err)
	}

	grpcSrv := grpcserver.NewGRPCServer(logg, calendar)
	go func() {
		logg.Info("grpc listening on %s", grpcLis.Addr().String())
		if serveErr := grpcSrv.Serve(grpcLis); serveErr != nil {
			logg.Error("grpc serve: %s", serveErr.Error())
		}
	}()

	httpServer := internalhttp.NewServer(config.HTTPServer.Host, config.HTTPServer.Port, logg, calendar)

	logg.Info("calendar is running (http+gRPC)...")
	// logg.Error("test error message...")
	// logg.Warning("test warning message...")
	// logg.Debug("test debug message...")

	if err := httpServer.Run(ctx); err != nil {
		logg.Error("http server: %s", err.Error())
		return fmt.Errorf("http server: %w", err)
	}

	grpcSrv.GracefulStop()
	return nil
}
