// client-cli is a test CLI that subscribes to order status updates via gRPC streaming.
// Usage:
//
//	go run main.go -addr localhost:50052 -order <order_id>
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Исправлен путь к сгенерированному коду заказов
	pb "github.com/yerdembek/AP2_assignment2/generated/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50052", "Order Service gRPC address")
	orderID := flag.String("order", "", "Order ID to subscribe to (required)")
	flag.Parse()

	if *orderID == "" {
		log.Fatal("Flag -order is required. Example: go run main.go -order ord_123456")
	}

	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Order Service at %s: %v", *addr, err)
	}
	defer conn.Close()

	client := pb.NewOrderTrackingServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Received signal, closing stream...")
		cancel()
	}()

	stream, err := client.SubscribeToOrderUpdates(ctx, &pb.OrderRequest{OrderId: *orderID})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Printf("Subscribed to order %s. Waiting for updates (press Ctrl+C to stop)...\n", *orderID)

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by server.")
			return
		}
		if err != nil {
			log.Printf("Stream error: %v", err)
			return
		}
		log.Printf("[UPDATE] order_id=%-20s  status=%-15s  time=%s  msg=%s",
			update.OrderId,
			update.Status,
			update.UpdatedAt.AsTime().Format("15:04:05"),
			update.Message,
		)
	}
}
