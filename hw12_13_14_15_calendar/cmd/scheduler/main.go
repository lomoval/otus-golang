package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/rabbit"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storagebuilder"
	log "github.com/sirupsen/logrus"
)

var configFile string

const (
	removeTimeout = time.Minute * 5
	checkTimout   = time.Minute
)

func newMessage(event storage.Event) rabbit.Message {
	return rabbit.Message{
		ID:      event.ID,
		Name:    event.Title,
		Time:    event.StartTime,
		OwnerID: event.OwnerID,
	}
}

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

	stor, err := storagebuilder.NewStorage(config.Storage)
	if err != nil {
		log.Errorf("failed to start %v", err)
		return
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		stor.Close(ctx)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	startTime := time.Now().Add(-time.Minute)
	endTime := time.Now()
	checkTicker := time.NewTicker(checkTimout)
	removeTicker := time.NewTicker(removeTimeout)
	for {
		select {
		case <-ctx.Done():
			return
		default:

			log.Debugf("get events: %s - %s", startTime, endTime)
			events, err := stor.GetEventsByNotifier(ctx, startTime, endTime)
			if err != nil {
				log.Errorf("failed to get events: %s", err)
				continue
			}
			for _, event := range events {
				log.Debugf("send event: %v", event)
				m := newMessage(event)
				data, _ := json.Marshal(m)
				r.Publish(data)
			}
			select {
			case <-ctx.Done():
				return
			case <-checkTicker.C:
				log.Debug("ticker")
				startTime = endTime
				endTime = time.Now()
			case <-removeTicker.C:
				stor.RemoveAfter(ctx, time.Now().Add(-1*(time.Hour*24*365)))
			}
		}
	}
}
