package logger

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Level string
}

func PrepareLogger(config Config) error {
	level, err := log.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("fialed to parse logger levev: %w", err)
	}
	log.SetLevel(level)
	return nil
}
