package storage

import (
	"time"
)

type Event struct {
	ID           string
	Title        string
	StartTime    time.Time
	EndTime      time.Time
	Description  string
	OwnerID      string
	NotifyBefore time.Duration
}
