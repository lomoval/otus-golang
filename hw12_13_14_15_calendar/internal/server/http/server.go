package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Server struct { // TODO
	srv *http.Server
}

type Application interface { // TODO
}

func NewServer(host string, port int, app Application) *Server {
	return &Server{srv: &http.Server{Addr: net.JoinHostPort(host, strconv.Itoa(port))}}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELLO !!!"))
	})
	s.srv.Handler = loggingMiddleware(mux)

	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Stop(c)
	}()

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
