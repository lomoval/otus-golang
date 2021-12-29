package internalgrpc

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func loggingHandler(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("method %q failed: %s", info.FullMethod, err)
	}
	ip := ""
	if peer, ok := peer.FromContext(ctx); ok {
		ip = peer.Addr.String()
	}
	var userAgent []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userAgent = md.Get("user-agent")
	}
	log.WithField("ip", ip).
		WithField("method", info.FullMethod).
		WithField("user-agent", userAgent).
		WithField("latency", time.Since(start)).
		Info("GRPC request processed")
	return resp, err
}
