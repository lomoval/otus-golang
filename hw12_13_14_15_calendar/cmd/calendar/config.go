package main

import (
	"fmt"
	"strings"

	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/server/http"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storagebuilder"
	"github.com/spf13/viper"
)

const envConfigPrefix = "$env:"

type Config struct {
	Server  internalhttp.Config
	Logger  logger.Config
	Storage storagebuilder.Config
}

func NewConfig(configFile string) (Config, error) {
	config := Config{}
	viper.SetConfigFile(configFile)

	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", "8005")
	viper.SetDefault("logger.level", "WARN")
	viper.SetDefault("storage.storageType", "memory")

	err := viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("failed to read config %q: %w", configFile, err)
	}
	keys := viper.AllKeys()
	for _, key := range keys {
		env := viper.GetString(key)
		if strings.HasPrefix(env, envConfigPrefix) {
			err := viper.BindEnv(key, env[len(envConfigPrefix):])
			if err != nil {
				return Config{}, fmt.Errorf("failed to prepare config: %w", err)
			}
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("unable to decode into config struct: %w", err)
	}
	return config, nil
}
