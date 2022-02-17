package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/rabbit"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/scheduler_config.yaml", "Path to configuration file")
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {
	flag.Parse()

	config, err := NewConfig(configFile)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}
	err = logger.PrepareLogger(config.Logger)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}

	r := rabbit.New(config.Rabbit)
	r.Connect()
	defer r.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	r.Consume(ctx, func(msg amqp.Delivery) {
		m := rabbit.Message{}
		err := json.Unmarshal(msg.Body, &m)
		if err != nil {
			log.Errorf("failed to parse bytes: %s", err)
			cancel()
			return
		}
		log.Printf("sending message %v", m)
	})
}
