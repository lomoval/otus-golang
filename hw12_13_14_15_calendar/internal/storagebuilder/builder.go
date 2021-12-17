package storagebuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/sql"
)

type Config struct {
	StorageType string
	Database    sqlstorage.Config
}

func New(config Config) (storage.Storage, error) {
	switch config.StorageType {
	case "memory":
		return memorystorage.New(), nil
	case "sql":
		s := sqlstorage.New(config.Database)
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
