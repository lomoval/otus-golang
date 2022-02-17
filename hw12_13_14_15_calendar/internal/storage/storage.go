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
	AddEvent(ctx context.Context, e *Event) error
	UpdateEvent(ctx context.Context, id string, e Event) error
	RemoveEvent(ctx context.Context, id string) error
	GetEventsForDay(ctx context.Context, date time.Time) ([]Event, error)
	GetEventsForWeek(ctx context.Context, startDate time.Time) ([]Event, error)
	GetEventsForMonth(ctx context.Context, startDate time.Time) ([]Event, error)
	GetEventsByNotifier(ctx context.Context, startTime time.Time, endTime time.Time) ([]Event, error)
	RemoveAfter(ctx context.Context, time time.Time) error
}
