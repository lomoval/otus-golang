package storage

import (
	"time"
)

type Event struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Description  string    `json:"description"`
	OwnerID      string    `json:"ownerId"`
	NotifyBefore int32     `json:"notifyBefore"`
}
