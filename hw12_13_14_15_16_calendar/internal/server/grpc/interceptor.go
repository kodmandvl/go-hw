package grpcserver

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// UnaryLoggingInterceptor — аналог HTTP-middleware: фиксируем метод, задержку и адрес клиента.
func UnaryLoggingInterceptor(log Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		latency := time.Since(start)

		client := "unknown"
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			client = p.Addr.String()
		}

		if err != nil {
			log.Error("grpc request %s from %s failed in %s: %v", info.FullMethod, client, latency, err)
			return resp, err
		}

		log.Info("grpc request %s from %s ok in %s", info.FullMethod, client, latency)
		return resp, nil
	}
}
