package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config для планировщика: БД (как у API), RabbitMQ и период тика.
type Config struct {
	Logger    LoggerConf    `mapstructure:"logger"`
	DB        DBConf        `mapstructure:"db"`
	Rabbit    RabbitConf    `mapstructure:"rabbit"`
	Scheduler SchedulerConf `mapstructure:"scheduler"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type DBConf struct {
	DBHost     string `mapstructure:"dbhost"`
	DBPort     int    `mapstructure:"dbport"`
	DBName     string `mapstructure:"dbname"`
	DBUsername string `mapstructure:"dbusername"`
	DBPassword string `mapstructure:"dbpassword"`
}

type RabbitConf struct {
	URI   string `mapstructure:"uri"`
	Queue string `mapstructure:"queue"`
}

type SchedulerConf struct {
	IntervalSeconds int `mapstructure:"interval_seconds"`
}

func NewConfig() *Config {
	vpr := viper.New()
	vpr.SetConfigFile(configFile)

	if err := vpr.ReadInConfig(); err != nil {
		fmt.Printf("load config: %s", err)
		os.Exit(1)
	}

	var config Config

	if err := vpr.Unmarshal(&config); err != nil {
		fmt.Printf("read config: %s", err)
		os.Exit(1)
	}

	return &config
}
