package logger

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor for access logging.
func UnaryServerInterceptor(l *Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		l.GRPCAccessLog(info.FullMethod, duration, err)
		return resp, err
	}
}
