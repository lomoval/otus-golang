package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/api"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/server/http"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	sqlstorage "github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storagebuilder"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	httpServerHost = "127.0.0.1"
	httpServerPort = 9005
	grpcServerHost = "127.0.0.1"
	grpcServerPort = 9006
	pgHost         = "127.0.0.1"
	pgPort         = 5432
	pgDatabase     = "testing"
	pgUsername     = "postgres"
	pgPassword     = "pas"
	storageType    = "memory"
	grpcGatewayURL = ""
	httpServerURL  = ""
)

func TestMain(m *testing.M) {
	logger.PrepareLogger(logger.Config{Level: "ERROR"})

	port := os.Getenv("TEST_HTTP_SERVER_PORT")
	if port != "" {
		httpServerPort, _ = strconv.Atoi(port)
	}
	port = os.Getenv("TEST_GRPC_SERVER_PORT")
	if port != "" {
		grpcServerPort, _ = strconv.Atoi(port)
	}

	host := os.Getenv("TEST_POSTGRES_HOST")
	if host != "" {
		pgHost = host
	}
	port = os.Getenv("TEST_POSTGRES_PORT")
	if port != "" {
		var err error
		pgPort, err = strconv.Atoi(port)
		if err != nil {
			log.Printf("failed to parse port '%s': %v", port, err)
			os.Exit(-1)
		}
	}

	storage := os.Getenv("TEST_STORAGE_TYPE")
	if storage != "" {
		storageType = storage
	}

	grpcGatewayURL = fmt.Sprintf("http://%s/Events/", net.JoinHostPort(httpServerHost, strconv.Itoa(httpServerPort)))
	httpServerURL = fmt.Sprintf("http://%s/", net.JoinHostPort(httpServerHost, strconv.Itoa(httpServerPort)))

	cleanupDB()
	code := m.Run()
	os.Exit(code)
}

// Wrapper to have own marshalling for duration type.
type testEvent struct {
	storage.Event
	NotifyBefore int32 `json:"notifyBefore"`
}

// For marshaling/unmarshalling JSON.
type apiStruct struct {
	ID     string      `json:"id,omitempty"`
	Event  testEvent   `json:"event"`
	Events []testEvent `json:"events"`
}

func TestStorage(t *testing.T) {
	t.Run("add event", func(t *testing.T) {
		startServer(t)

		event := createEvent()
		jsonStr, err := json.Marshal(apiStruct{Event: event})
		require.NoError(t, err)

		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", jsonStr)
		defer resp.Body.Close()

		require.Equal(t, 200, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse body")
		require.NotEmpty(t, actual.Event.ID)

		actual.Event.ID = ""
		compareEvents(t, event, actual.Event)
	})

	t.Run("update get event", func(t *testing.T) {
		startServer(t)

		event := createEvent()
		jsonStr, err := json.Marshal(apiStruct{Event: event})
		require.NoError(t, err)

		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", jsonStr)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var expected apiStruct
		require.NoError(t, json.Unmarshal(body, &expected), "failed to parse response")
		expected.ID = expected.Event.ID
		expected.Event.Title += "UPD"
		expected.Event.Description += "UPD"
		expected.Event.NotifyBefore += 5
		expected.Event.StartTime = expected.Event.StartTime.Add(1 * time.Minute)
		expected.Event.EndTime = expected.Event.EndTime.Add(1 * time.Minute)

		jsonStr, err = json.Marshal(expected)
		require.NoError(t, err)
		updResp := sendRequest(t, "POST", grpcGatewayURL, "UpdateEvent", jsonStr)
		defer updResp.Body.Close()

		require.Equal(t, 200, updResp.StatusCode)
		body, err = ioutil.ReadAll(updResp.Body)
		require.NoError(t, err, "failed to read body")
		require.Equal(t, string(body), "{}")

		getResp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForDay",
			[]byte(`{"startDate": "`+expected.Event.StartTime.Local().Format(time.RFC3339)+`"}`),
		)
		defer getResp.Body.Close()
		require.Equal(t, 200, getResp.StatusCode)
		body, err = ioutil.ReadAll(getResp.Body)
		require.NoError(t, err, "failed to read body")
		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 1, len(actual.Events))
		compareEvents(t, expected.Event, actual.Events[0])
	})

	t.Run("remove event", func(t *testing.T) {
		startServer(t)

		event := createEvent()
		jsonStr, err := json.Marshal(apiStruct{Event: event})
		require.NoError(t, err)

		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", jsonStr)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var got apiStruct
		require.NoError(t, json.Unmarshal(body, &got), "failed to parse response")

		updResp := sendRequest(t, "POST", grpcGatewayURL, "RemoveEvent", []byte(`{"id": "`+got.Event.ID+`"}`))
		defer updResp.Body.Close()

		require.Equal(t, 200, updResp.StatusCode)
		body, err = ioutil.ReadAll(updResp.Body)
		require.NoError(t, err, "failed to read body")
		require.Equal(t, string(body), "{}")

		getResp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForDay",
			[]byte(`{"startDate": "`+got.Event.StartTime.Local().Format(time.RFC3339)+`"}`),
		)
		defer getResp.Body.Close()
		require.Equal(t, 200, getResp.StatusCode)
		body, err = ioutil.ReadAll(getResp.Body)
		require.NoError(t, err, "failed to read body")
		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 0, len(actual.Events))
	})
}

func TestGatewayGetEvents(t *testing.T) {
	startServer(t)

	initDate := time.Date(2300, 0o1, 0o1, 0, 0, 0, 0, time.UTC)
	event := createEvent()
	event.StartTime = initDate
	event.EndTime = initDate.Add(2 * time.Hour)
	events := make([]testEvent, 0, 60)

	for i := 0; i < 60; i++ {
		jsonStr, err := json.Marshal(apiStruct{Event: event})
		require.NoError(t, err)
		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", jsonStr)
		require.Equal(t, 200, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err, "failed to read body")

		var got apiStruct
		require.NoError(t, json.Unmarshal(body, &got), "failed to parse response")
		events = append(events, got.Event)

		event.Title += strconv.Itoa(i)
		event.StartTime = event.StartTime.AddDate(0, 0, 1)
		event.EndTime = event.EndTime.AddDate(0, 0, 1)
	}

	t.Run("get day", func(t *testing.T) {
		resp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForDay",
			[]byte(`{"startDate": "`+initDate.Local().Format(time.RFC3339)+`"}`),
		)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 1, len(actual.Events))
		compareEvents(t, events[0], actual.Events[0])
	})

	t.Run("get week", func(t *testing.T) {
		resp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForWeek",
			[]byte(`{"startDate": "`+initDate.Local().Format(time.RFC3339)+`"}`),
		)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 7, len(actual.Events))
		sort.Slice(actual.Events, func(i, j int) bool {
			return actual.Events[i].StartTime.Before(actual.Events[j].StartTime)
		})

		for i := 0; i < 7; i++ {
			compareEvents(t, events[i], actual.Events[i])
		}
	})

	t.Run("get month", func(t *testing.T) {
		resp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForMonth",
			[]byte(`{"startDate": "`+initDate.Local().Format(time.RFC3339)+`"}`),
		)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 31, len(actual.Events))
		sort.Slice(actual.Events, func(i, j int) bool {
			return actual.Events[i].StartTime.Before(actual.Events[j].StartTime)
		})

		for i := 0; i < 31; i++ {
			compareEvents(t, events[i], actual.Events[i])
		}
	})

	t.Run("get month 28 days", func(t *testing.T) {
		resp := sendRequest(
			t,
			"POST",
			grpcGatewayURL,
			"GetEventsForMonth",
			[]byte(`{"startDate": "`+initDate.AddDate(0, 1, 0).Local().Format(time.RFC3339)+`"}`),
		)
		defer resp.Body.Close()
		require.Equal(t, 200, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read body")

		var actual apiStruct
		require.NoError(t, json.Unmarshal(body, &actual), "failed to parse response")
		require.Equal(t, 28, len(actual.Events))
		sort.Slice(actual.Events, func(i, j int) bool {
			return actual.Events[i].StartTime.Before(actual.Events[j].StartTime)
		})

		for i := 0; i < 28; i++ {
			compareEvents(t, events[i+31], actual.Events[i])
		}
	})
}

func TestGatewayErrors(t *testing.T) {
	startServer(t)

	t.Run("add no empty event", func(t *testing.T) {
		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", []byte(`{"event": {}}`))
		defer resp.Body.Close()
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("add no event", func(t *testing.T) {
		resp := sendRequest(t, "POST", grpcGatewayURL, "AddEvent", []byte(`{}`))
		defer resp.Body.Close()
		require.Equal(t, 400, resp.StatusCode)
	})

	t.Run("remove non exists event", func(t *testing.T) {
		resp := sendRequest(t, "POST", grpcGatewayURL, "RemoveEvent", []byte(`{"id": "_non_exists_"}`))
		defer resp.Body.Close()
		require.Equal(t, 404, resp.StatusCode)
	})

	t.Run("update non exists event", func(t *testing.T) {
		event := createEvent()
		jsonStr, err := json.Marshal(apiStruct{ID: "__non_exist__", Event: event})
		require.NoError(t, err)

		resp := sendRequest(t, "POST", grpcGatewayURL, "UpdateEvent", jsonStr)
		defer resp.Body.Close()
		require.Equal(t, 404, resp.StatusCode)
	})
}

func sendRequest(t *testing.T, method string, url string, path string, requestBody []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		method,
		url+path,
		bytes.NewBuffer(requestBody),
	)
	require.NoError(t, err, "failed to send request")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		require.NoError(t, err, "failed send request")
	}
	return resp
}

func startServer(t *testing.T) {
	t.Helper()

	storage, err := storagebuilder.NewStorage(storagebuilder.Config{
		StorageType: storageType,
		Database: sqlstorage.Config{
			Host:     pgHost,
			Port:     pgPort,
			Database: pgDatabase,
			Username: pgUsername,
			Password: pgPassword,
		},
	})
	require.NoError(t, err, "failed to create storage")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, storage.Connect(ctx))

	calendar := app.New(storage)
	httpServer := internalhttp.NewServer(internalhttp.Config{
		Host: httpServerHost,
		Port: httpServerPort,
	}, calendar)
	grpcServer := internalgrpc.NewServer(internalgrpc.Config{
		Host: grpcServerHost,
		Port: grpcServerPort,
	}, calendar)

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		grpcServer.Start(ctx)
	}()

	// Wait stating of servers
	require.Eventually(t, func() bool {
		conn, err := grpc.Dial(
			net.JoinHostPort(grpcServerHost, strconv.Itoa(grpcServerPort)),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err, "failed to dial to grpc server")
		client := api.NewEventsClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err = client.GetEventsForDay(ctx, &api.GetEventsRequest{StartDate: timestamppb.Now()})
		return err == nil
	}, 5*time.Second, 200*time.Millisecond)

	gatewayMux, err := grpcServer.GatewayMux(ctx)
	if err != nil {
		require.NoError(t, err, "failed to get gateway mux")
	}

	go func() {
		httpServer.Start(ctx, gatewayMux)
	}()

	require.Eventually(t, func() bool {
		resp := sendRequest(t, "GET", httpServerURL, "hello", nil)
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}, 5*time.Second, 200*time.Millisecond)

	t.Cleanup(func() {
		cancel()
		grpcServer.Stop(ctx)
		httpServer.Stop(ctx)
		storage.Close(ctx)
		require.NoError(t, cleanupDB())
	})
}

func createEvent() testEvent {
	return testEvent{
		Event: storage.Event{
			ID:          "",
			Title:       "Test",
			StartTime:   time.Now().Truncate(time.Second).Add(5 * time.Minute),
			EndTime:     time.Now().Truncate(time.Second).Add(20 * time.Minute),
			Description: "TestDescription",
			OwnerID:     "OwnId",
		},
		NotifyBefore: 1,
	}
}

func compareEvents(t *testing.T, expected testEvent, actual testEvent) {
	t.Helper()
	require.True(
		t,
		expected.StartTime.Equal(actual.StartTime),
		"start time is not equals %q != %q", expected.StartTime, actual.StartTime)
	require.True(
		t,
		expected.StartTime.Equal(actual.StartTime),
		"start time is not equals %q != %q", expected.StartTime, actual.StartTime)
	expected.StartTime = actual.StartTime
	expected.EndTime = actual.EndTime
	require.Equal(t, expected, actual)
}

func cleanupDB() error {
	if storageType != "sql" {
		return nil
	}
	db, err := sqlx.Connect(
		"postgres",
		fmt.Sprintf(
			"sslmode=disable host=%s port=%d dbname=%s user=%s password=%s",
			pgHost,
			pgPort,
			pgDatabase,
			pgUsername,
			pgPassword,
		),
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
