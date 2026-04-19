package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config для рассыльщика: только лог и RabbitMQ.
type Config struct {
	Logger LoggerConf `mapstructure:"logger"`
	Rabbit RabbitConf `mapstructure:"rabbit"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type RabbitConf struct {
	URI   string `mapstructure:"uri"`
	Queue string `mapstructure:"queue"`
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
