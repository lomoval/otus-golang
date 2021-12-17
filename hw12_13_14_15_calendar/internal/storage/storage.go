package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicateEventID   = errors.New("event with same ID exists")
	ErrNotFoundEvent      = errors.New("event not found")
	ErrIncorrectStartDate = errors.New("date should be a first day of requested period")
	ErrIncorrectEventTime = errors.New("incorrect event time")
)

type Storage interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	AddEvent(e *Event) error
	UpdateEvent(id string, e Event) error
	RemoveEvent(id string) error
	GetEventsForDay(date time.Time) ([]Event, error)
	GetEventsForWeek(startDate time.Time) ([]Event, error)
	GetEventsForMonth(startDate time.Time) ([]Event, error)
}
