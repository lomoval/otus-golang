package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lomoval/otus-golang/hw12_13_14_15_calendar/internal/app"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	srv  *http.Server
	addr string
}

func NewServer(config Config, app *app.App) *Server {
	return &Server{
		addr: net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
		srv:  &http.Server{Addr: net.JoinHostPort(config.Host, strconv.Itoa(config.Port))},
	}
}

func (s *Server) Start(_ context.Context, mux *runtime.ServeMux) error {
	if mux == nil {
		mux = runtime.NewServeMux()
	}

	mux.HandlePath("GET", "/hello", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.Write([]byte("HELLO !!!"))
	})
	s.srv.Handler = loggingMiddleware(mux)

	log.Printf("starting http server on %s", s.addr)
	err := s.srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server failed: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func getIP(req *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	if parsed := net.ParseIP(ip); parsed == nil {
		return "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return ip, nil
}

// TODO
