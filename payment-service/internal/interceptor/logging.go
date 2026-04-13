package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// UnaryLoggingInterceptor logs each incoming unary RPC call with method name and duration.
// This is the bonus interceptor (middleware) requirement.
func UnaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	log.Printf("[gRPC] --> %s | request: %+v", info.FullMethod, req)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		log.Printf("[gRPC] <-- %s | duration: %s | ERROR: %v", info.FullMethod, duration, err)
	} else {
		log.Printf("[gRPC] <-- %s | duration: %s | OK", info.FullMethod, duration)
	}
	return resp, err
}
