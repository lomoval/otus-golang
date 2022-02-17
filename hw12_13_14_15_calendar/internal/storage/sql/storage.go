package sqlstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/util"
	log "github.com/sirupsen/logrus"
)

var ErrConnectionFailed = errors.New("failed to connect")

const dbErrUniqueViolation = "23505"

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type Storage struct {
	host         string
	port         int
	database     string
	username     string
	password     string
	db           *sqlx.DB
	firstWeekDay time.Weekday
}

func New(config Config) *Storage {
	return &Storage{
		host:         config.Host,
		port:         config.Port,
		database:     config.Database,
		username:     config.Username,
		password:     config.Password,
		firstWeekDay: time.Monday,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(
		ctx,
		"postgres",
		fmt.Sprintf(
			"sslmode=disable host=%s port=%d dbname=%s user=%s password=%s",
			s.host, s.port, s.database, s.username, s.password),
	)
	if err != nil {
		log.Errorf("failed to connect: %v", err)
		return ErrConnectionFailed
	}
	s.db = db
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

func (s *Storage) AddEvent(ctx context.Context, e *storage.Event) error {
	if e.StartTime.Before(time.Now()) {
		return fmt.Errorf("start time of the event must be in the future: %w", storage.ErrIncorrectEventTime)
	}
	if !e.EndTime.After(e.StartTime) {
		return fmt.Errorf("event end time should be after of start time: %w", storage.ErrIncorrectEventTime)
	}

	var err error
	switch e.ID {
	case "":
		err = s.db.GetContext(
			ctx,
			&e.ID,
			"INSERT INTO Events(title, start_timestamp, end_timestamp, description, notify_before, owner_id) "+
				"VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
			e.Title, e.StartTime.UTC(), e.EndTime.UTC(), e.Description, e.NotifyBefore, e.OwnerID)
	default:
		_, err = s.db.ExecContext(
			ctx,
			"INSERT INTO Events(id, title, start_timestamp, end_timestamp, description, notify_before, owner_id) "+
				"VALUES($1, $2, $3, $4, $5, $6, $7)",
			e.ID, e.Title, e.StartTime.UTC(), e.EndTime.UTC(), e.Description, e.NotifyBefore, e.OwnerID)
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == dbErrUniqueViolation {
		return fmt.Errorf("duplicate ID %q: %w", e.ID, storage.ErrDuplicateEventID)
	}

	return err
}

func (s *Storage) UpdateEvent(ctx context.Context, id string, e storage.Event) error {
	if e.StartTime.Before(time.Now()) {
		return fmt.Errorf("start time of the event must be in the future: %w", storage.ErrIncorrectEventTime)
	}
	if !e.EndTime.After(e.StartTime) {
		return fmt.Errorf("event end time should be after of start time: %w", storage.ErrIncorrectEventTime)
	}

	var found bool
	err := s.db.GetContext(
		ctx,
		&found,
		"UPDATE Events SET title=$2, start_timestamp=$3, end_timestamp=$4, description=$5, notify_before=$6 "+
			"WHERE id=$1 RETURNING TRUE",
		id,
		e.Title,
		e.StartTime,
		e.EndTime,
		e.Description,
		e.NotifyBefore,
	)

	if !found {
		return fmt.Errorf("failed to update event with id %q: %w", e.ID, storage.ErrNotFoundEvent)
	}
	return err
}

func (s *Storage) RemoveEvent(ctx context.Context, id string) error {
	var found bool
	err := s.db.GetContext(ctx, &found, "DELETE FROM Events WHERE id=$1 RETURNING TRUE", id)

	if !found {
		return fmt.Errorf("failed to remove event with id %q: %w", id, storage.ErrNotFoundEvent)
	}
	return err
}

func (s *Storage) GetEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endTime := startTime.Add(24 * time.Hour)
	return s.selectByRange(ctx, startTime, endTime)
}

func (s *Storage) GetEventsForWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startTime := util.TruncateToDay(startDate)
	if startTime.Weekday() != s.firstWeekDay {
		return nil, storage.ErrIncorrectStartDate
	}
	endTime := startTime.AddDate(0, 0, 7)
	return s.selectByRange(ctx, startTime, endTime)
}

func (s *Storage) GetEventsForMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	startTime := util.TruncateToDay(startDate)
	if startTime.Day() != 1 {
		return nil, storage.ErrIncorrectStartDate
	}
	endTime := startTime.AddDate(0, 1, 0)
	return s.selectByRange(ctx, startTime, endTime)
}

func (s *Storage) GetEventsByNotifier(
	ctx context.Context,
	startTime time.Time,
	endTime time.Time,
) ([]storage.Event, error) {
	var events []storage.Event
	err := s.db.SelectContext(
		ctx,
		&events,
		"SELECT id, title, start_timestamp AS startTime, end_timestamp AS endTime, description, "+
			"notify_before AS notifyBefore, owner_id AS ownerId "+
			"FROM Events WHERE notify_before > 0 AND (start_timestamp - (interval '1' day * notify_before))>=$1 "+
			"AND (start_timestamp - (interval '1' day * notify_before))<=$2",
		startTime,
		endTime,
	)

	return events, err
}

func (s *Storage) RemoveAfter(ctx context.Context, time time.Time) error {
	_, err := s.db.Exec("DELETE FROM Events WHERE start_timestamp < $1", time)
	return err
}

// Select in range [startTime:endTime).
func (s *Storage) selectByRange(ctx context.Context, startTime time.Time, endTime time.Time) ([]storage.Event, error) {
	var events []storage.Event
	err := s.db.SelectContext(
		ctx,
		&events,
		"SELECT id, title, start_timestamp AS startTime, end_timestamp AS endTime, description, "+
			"notify_before AS notifyBefore, owner_id AS ownerId "+
			"FROM Events WHERE start_timestamp>=$1 AND end_timestamp<$2",
		startTime,
		endTime,
	)

	return events, err
}
