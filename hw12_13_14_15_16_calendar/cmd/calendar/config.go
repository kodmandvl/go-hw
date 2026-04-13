package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger     LoggerConf     `mapstructure:"logger"`
	Storage    StorageConf    `mapstructure:"storage"`
	DB         DBConf         `mapstructure:"db"`
	HTTPServer HTTPServerConf `mapstructure:"http"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type StorageConf struct {
	Type string `mapstructure:"type"`
}

type DBConf struct {
	DBHost     string `mapstructure:"dbhost"`
	DBPort     int    `mapstructure:"dbport"`
	DBName     string `mapstructure:"dbname"`
	DBUsername string `mapstructure:"dbusername"`
	DBPassword string `mapstructure:"dbpassword"`
}

type HTTPServerConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func NewConfig() *Config {
	// Для YAML используем Viper.
	vpr := viper.New()
	vpr.SetConfigFile(configFile)

	// ReadInConfig will discover and load the configuration file.
	if err := vpr.ReadInConfig(); err != nil {
		fmt.Printf("load config: %s", err)
		os.Exit(1)
	}

	var config Config

	// Unmarshal unmarshals the config into a Struct.
	if err := vpr.Unmarshal(&config); err != nil {
		fmt.Printf("read config: %s", err)
	}

	return &config
}
