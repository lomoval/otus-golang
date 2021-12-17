package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

type Config struct {
	Host string
	Port int
}

type Server struct { // TODO
	srv *http.Server
}

type Application interface { // TODO
}

func NewServer(config Config, app Application) *Server {
	return &Server{srv: &http.Server{Addr: net.JoinHostPort(config.Host, strconv.Itoa(config.Port))}}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELLO !!!"))
	})
	s.srv.Handler = loggingMiddleware(mux)

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
