// +build sql

package sqlstorage_test

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	sqlstorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	host     = "127.0.0.1"
	port     = 5532
	database = "testing"
	username = "postgres"
	password = "pas"
)

func TestMain(m *testing.M) {
	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	if pgHost != "" {
		host = pgHost
	}
	if pgPort != "" {
		port, _ = strconv.Atoi(pgPort)
	}

	cleanupDb()
	code := m.Run()
	os.Exit(code)
}

func TestStorage(t *testing.T) {
	t.Run("add event", func(t *testing.T) {
		initDate := time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC)
		e := storage.Event{
			ID:           "",
			Title:        "test",
			StartTime:    initDate.Add(1 * time.Hour),
			EndTime:      initDate.Add(2 * time.Hour),
			Description:  "description",
			OwnerID:      "testId",
			NotifyBefore: 0,
		}
		s := createStorage(t)

		require.NoError(t, s.AddEvent(&e))
		require.NotEmpty(t, e.ID)

		events, err := s.GetEventsForDay(initDate)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		compareEvents(t, e, events[0])
	})

	t.Run("update event", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{
			ID:           "",
			Title:        "test",
			StartTime:    initDate.Add(1 * time.Hour),
			EndTime:      initDate.Add(2 * time.Hour),
			Description:  "description",
			OwnerID:      "testId",
			NotifyBefore: 0,
		}

		s := createStorage(t)
		require.NoError(t, s.AddEvent(&e))

		e.Title = "updated title"
		e.StartTime = e.EndTime.Add(21 * time.Minute)
		e.EndTime = e.EndTime.Add(33 * time.Minute)
		e.Description = "updated description"
		e.NotifyBefore = 100

		require.NoError(t, s.UpdateEvent(e.ID, e))

		events, err := s.GetEventsForWeek(initDate)
		require.NoError(t, err)
		require.Equal(t, 1, len(events))
		compareEvents(t, e, events[0])
	})

	t.Run("delete event", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{
			ID:           "",
			Title:        "test",
			StartTime:    initDate.Add(1 * time.Hour),
			EndTime:      initDate.Add(2 * time.Hour),
			Description:  "description",
			OwnerID:      "testId",
			NotifyBefore: 0,
		}

		s := createStorage(t)
		require.NoError(t, s.AddEvent(&e))

		require.NoError(t, s.RemoveEvent(e.ID))

		events, err := s.GetEventsForWeek(initDate)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("list", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{
			ID:           "",
			Title:        "test",
			StartTime:    initDate,
			EndTime:      initDate.Add(2 * time.Hour),
			Description:  "description",
			OwnerID:      "testId",
			NotifyBefore: 0,
		}

		s := createStorage(t)

		for i := 0; i < 60; i++ {
			require.NoError(t, s.AddEvent(&e))
			e.ID = ""
			e.StartTime = e.StartTime.AddDate(0, 0, 1)
			e.EndTime = e.EndTime.AddDate(0, 0, 1)
		}

		list, err := s.GetEventsForDay(initDate)
		require.NoError(t, err)
		require.Equal(t, len(list), 1)

		list, err = s.GetEventsForWeek(initDate)
		require.NoError(t, err)
		require.Equal(t, len(list), 7)

		list, err = s.GetEventsForMonth(initDate)
		require.NoError(t, err)
		require.Equal(t, len(list), 31)

		list, err = s.GetEventsForMonth(initDate.AddDate(0, 1, 0))
		require.NoError(t, err)
		require.Equal(t, len(list), 28)
	})
}

func TestStorageNegativeCases(t *testing.T) {
	t.Run("add event with same id", func(t *testing.T) {
		initDate := time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC)
		e := storage.Event{
			ID:           "",
			Title:        "test",
			StartTime:    initDate.Add(1 * time.Hour),
			EndTime:      initDate.Add(2 * time.Hour),
			Description:  "description",
			OwnerID:      "testId",
			NotifyBefore: 0,
		}
		s := createStorage(t)

		require.NoError(t, s.AddEvent(&e))
		require.ErrorIs(t, s.AddEvent(&e), storage.ErrDuplicateEventID)
	})

	t.Run("update not exist event", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{ID: "___not_exists___", StartTime: initDate, EndTime: initDate.Add(time.Hour)}
		s := createStorage(t)

		require.ErrorIs(t, s.UpdateEvent(e.ID, e), storage.ErrNotFoundEvent)
	})

	t.Run("delete not exist event event", func(t *testing.T) {
		e := storage.Event{ID: "___not_exists___"}
		s := createStorage(t)

		require.ErrorIs(t, s.RemoveEvent(e.ID), storage.ErrNotFoundEvent)
	})

	t.Run("old event time for insert", func(t *testing.T) {
		initDate := time.Now().Add(-1 * time.Minute)
		e := storage.Event{StartTime: initDate.Add(time.Hour), EndTime: initDate}
		s := createStorage(t)

		require.ErrorIs(t, s.AddEvent(&e), storage.ErrIncorrectEventTime)
	})

	t.Run("old event time for update", func(t *testing.T) {
		initDate := time.Now().Add(-1 * time.Minute)
		e := storage.Event{StartTime: initDate.Add(time.Hour), EndTime: initDate}
		s := createStorage(t)

		require.ErrorIs(t, s.UpdateEvent(e.ID, e), storage.ErrIncorrectEventTime)
	})

	t.Run("incorrect event time for insert", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{StartTime: initDate.Add(time.Hour), EndTime: initDate}
		s := createStorage(t)

		require.ErrorIs(t, s.AddEvent(&e), storage.ErrIncorrectEventTime)
	})

	t.Run("incorrect event time for insert", func(t *testing.T) {
		initDate := time.Date(2300, 01, 01, 0, 0, 0, 0, time.UTC)
		e := storage.Event{StartTime: initDate.Add(time.Hour), EndTime: initDate}
		s := createStorage(t)

		require.ErrorIs(t, s.UpdateEvent(e.ID, e), storage.ErrIncorrectEventTime)
	})
}

func TestStorageValidateStarDates(t *testing.T) {
	tests := []struct {
		testFunc    func(s *sqlstorage.Storage) error
		expectedErr error
	}{
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForWeek(time.Date(2021, 12, 06, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: nil,
		},
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForWeek(time.Date(2300, 01, 8, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: nil,
		},
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForWeek(time.Date(2300, 01, 29, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: nil,
		},
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForMonth(time.Date(2300, 1, 1, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: nil,
		},
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForWeek(time.Date(2300, 01, 02, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: storage.ErrIncorrectStartDate,
		},
		{
			testFunc: func(s *sqlstorage.Storage) error {
				_, err := s.GetEventsForMonth(time.Date(2300, 01, 02, 0, 0, 0, 0, time.UTC))
				return err
			},
			expectedErr: storage.ErrIncorrectStartDate,
		},
	}

	s := createStorage(t)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			require.ErrorIs(t, tt.testFunc(s), tt.expectedErr)
		})
	}
}

func cleanupDb() error {
	db, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf("sslmode=disable host=%s port=%d dbname=%s user=%s password=%s", host, port, database, username, password),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE Events")
	if err != nil {
		return err
	}
	return err
}

func compareEvents(t *testing.T, expected storage.Event, actual storage.Event) {
	t.Helper()
	require.True(t, expected.StartTime.Equal(actual.StartTime), "start time is not equals %q != %q", expected.StartTime, actual.StartTime)
	require.True(t, expected.StartTime.Equal(actual.StartTime), "start time is not equals %q != %q", expected.StartTime, actual.StartTime)
	expected.StartTime = actual.StartTime
	expected.EndTime = actual.EndTime
	require.Equal(t, expected, actual)
}

func createStorage(t *testing.T) *sqlstorage.Storage {
	t.Helper()
	s := sqlstorage.New(host, port, database, username, password)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	require.NoError(t, s.Connect(ctx))
	t.Cleanup(func() {
		s.Close(ctx)
		require.NoError(t, cleanupDb())
	})
	return s
}
