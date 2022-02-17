//go:generate -command PROTOC protoc -I../../../api/proto ../../../api/proto/event.proto ../../../api/proto/service.proto
//go:generate PROTOC --go_out=../../../api/ --go-grpc_out=../../../api/
//go:generate PROTOC --grpc-gateway_out ../../../api/ --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true

package internalgrpc

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/api"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	errEventNotProvided    = "event is not provided"
	errInternalServerError = "internal server error"
	errEventNotFound       = "event not found"
	errIncorrectEventTime  = "incorrect event time"
	errIncorrectDate       = "incorrect date"
	errDateIsNotProvided   = "date is not provided"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	api.UnimplementedEventsServer
	grpcServer *grpc.Server
	app        *app.App
	addr       string
}

func NewServer(config Config, app *app.App) *Server {
	return &Server{app: app, addr: net.JoinHostPort(config.Host, strconv.Itoa(config.Port))}
}

func (s *Server) Start(_ context.Context) error {
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(loggingHandler))
	api.RegisterEventsServer(s.grpcServer, s)

	lsn, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Errorf("failed to listen grpc endpoint: %v", err)
		return err
	}

	log.Printf("starting grpc server on %s", s.addr)
	err = s.grpcServer.Serve(lsn)
	return err
}

func (s *Server) GatewayMux(ctx context.Context) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := api.RegisterEventsHandlerFromEndpoint(ctx, mux, s.addr, opts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func (s *Server) Stop(_ context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}

func (s *Server) AddEvent(ctx context.Context, r *api.AddEventRequest) (*api.AddEventResponse, error) {
	if r.Event == nil {
		return nil, status.Errorf(codes.InvalidArgument, errEventNotProvided)
	}
	event, err := toStorageEvent(r.GetEvent())
	if err != nil {
		if errors.Is(err, storage.ErrIncorrectEventTime) {
			return nil, status.Errorf(codes.InvalidArgument, errIncorrectEventTime)
		}
		log.Errorf("failed to convert events: %v", err)
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}

	event.ID, err = s.app.CreateEvent(ctx, event)
	if err != nil {
		return nil, err
	}
	return &api.AddEventResponse{Event: toAPIEvent(event)}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, r *api.UpdateEventRequest) (*empty.Empty, error) {
	if r.GetEvent() == nil {
		return nil, status.Errorf(codes.InvalidArgument, errEventNotProvided)
	}
	event, err := toStorageEvent(r.GetEvent())
	if err != nil {
		if errors.Is(err, storage.ErrIncorrectEventTime) {
			return nil, status.Errorf(codes.InvalidArgument, errIncorrectEventTime)
		}
		log.Errorf("failed to convert events: %v", err)
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}

	err = s.app.UpdateEvent(ctx, r.GetId(), event)
	if err != nil {
		if errors.Is(err, storage.ErrNotFoundEvent) {
			return nil, status.Errorf(codes.NotFound, errEventNotFound)
		}
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &empty.Empty{}, nil
}

func (s *Server) RemoveEvent(ctx context.Context, r *api.RemoveEventRequest) (*empty.Empty, error) {
	err := s.app.RemoveEvent(ctx, r.GetId())
	if err != nil {
		if errors.Is(err, storage.ErrNotFoundEvent) {
			return nil, status.Errorf(codes.NotFound, errEventNotFound)
		}
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Server) GetEventsForDay(ctx context.Context, r *api.GetEventsRequest) (*api.GetEventsResponse, error) {
	date := r.GetStartDate()
	if date == nil {
		return nil, status.Errorf(codes.InvalidArgument, errDateIsNotProvided)
	}
	if !date.IsValid() {
		return nil, status.Errorf(codes.InvalidArgument, errIncorrectDate)
	}
	events, err := s.app.GetEventsForDay(ctx, date.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}

	return &api.GetEventsResponse{Events: toAPIEvents(events)}, nil
}

func (s *Server) GetEventsForWeek(ctx context.Context, r *api.GetEventsRequest) (*api.GetEventsResponse, error) {
	date := r.GetStartDate()
	if date == nil {
		return nil, status.Errorf(codes.InvalidArgument, errDateIsNotProvided)
	}
	if !date.IsValid() {
		return nil, status.Errorf(codes.InvalidArgument, errIncorrectDate)
	}
	events, err := s.app.GetEventsForWeek(ctx, date.AsTime())
	if err != nil {
		if errors.Is(err, storage.ErrIncorrectStartDate) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}

	return &api.GetEventsResponse{Events: toAPIEvents(events)}, nil
}

func (s *Server) GetEventsForMonth(ctx context.Context, r *api.GetEventsRequest) (*api.GetEventsResponse, error) {
	date := r.GetStartDate()
	if date == nil {
		return nil, status.Errorf(codes.InvalidArgument, errDateIsNotProvided)
	}
	if !date.IsValid() {
		return nil, status.Errorf(codes.InvalidArgument, errIncorrectDate)
	}
	events, err := s.app.GetEventsForMonth(ctx, date.AsTime())
	if err != nil {
		if errors.Is(err, storage.ErrIncorrectStartDate) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}

	return &api.GetEventsResponse{Events: toAPIEvents(events)}, nil
}

func toStorageEvent(e *api.Event) (storage.Event, error) {
	if !e.StartTime.IsValid() || !e.EndTime.IsValid() {
		return storage.Event{}, storage.ErrIncorrectEventTime
	}
	return storage.Event{
		ID:           e.Id,
		Title:        e.Title,
		StartTime:    e.StartTime.AsTime(),
		EndTime:      e.EndTime.AsTime(),
		Description:  e.Description,
		OwnerID:      e.OwnerId,
		NotifyBefore: e.NotifyBefore,
	}, nil
}

func toAPIEvent(e storage.Event) *api.Event {
	return &api.Event{
		Id:           e.ID,
		Title:        e.Title,
		StartTime:    timestamppb.New(e.StartTime),
		EndTime:      timestamppb.New(e.EndTime),
		Description:  e.Description,
		OwnerId:      e.OwnerID,
		NotifyBefore: e.NotifyBefore,
	}
}

func toAPIEvents(events []storage.Event) []*api.Event {
	apiEvents := make([]*api.Event, 0, len(events))
	for _, event := range events {
		apiEvents = append(apiEvents, toAPIEvent(event))
	}
	return apiEvents
}
