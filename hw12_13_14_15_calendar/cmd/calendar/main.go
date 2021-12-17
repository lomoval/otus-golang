package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/app"
	internalhttp "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/server/http"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/sql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := prepareConfig(configFile)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}
	err = prepareLogger(config)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}
	err = prepareLogger(config)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}
	stor, err := prepareStorage(config)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}

	calendar := app.New(stor)

	server := internalhttp.NewServer(config.Host, config.Port, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			log.Error("failed to stop http server: " + err.Error())
		}
	}()

	log.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		log.Error("failed to start http server: " + err.Error())
		cancel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		err := stor.Close(ctx)
		if err != nil {
			log.Errorf("failed to close storage: %v", err)
		}
		os.Exit(1) //nolint:gocritic
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err = stor.Close(ctx)
	if err != nil {
		log.Errorf("failed to close storage: %v", err)
	}
}

func prepareConfig(configFile string) (Config, error) {
	config := NewConfig()
	viper.SetConfigFile(configFile)

	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", "8005")
	viper.SetDefault("logger.level", "WARN")
	viper.SetDefault("storage", "memory")

	err := viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("failed to read config %q: %w", configFile, err)
	}
	keys := viper.AllKeys()
	for _, key := range keys {
		env := viper.GetString(key)
		if strings.HasPrefix(env, "$env:") {
			err := viper.BindEnv(key, env[5:])
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

func prepareLogger(config Config) error {
	level, err := log.ParseLevel(config.Logger.Level)
	if err != nil {
		return fmt.Errorf("fialed to parse logger levev: %w", err)
	}
	log.SetLevel(level)
	return nil
}

func prepareStorage(config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "memory":
		return memorystorage.New(), nil
	case "sql":
		s := sqlstorage.New(config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			config.Database.Username,
			config.Database.Password,
		)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := s.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database %s %d: %w", config.Database.Host, config.Database.Port, err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unknown storage type %s", config.StorageType)
	}
}
