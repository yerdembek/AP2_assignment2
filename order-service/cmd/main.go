package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// Исправлен путь к сгенерированному коду
	pb "github.com/yerdembek/AP2_assignment2/generated/order"

	// Исправлены пути к внутренним пакетам сервиса
	"github.com/yerdembek/AP2_assignment2/order-service/internal/client"
	deliverygrpc "github.com/yerdembek/AP2_assignment2/order-service/internal/delivery/grpc"
	deliveryhttp "github.com/yerdembek/AP2_assignment2/order-service/internal/delivery/http"
	"github.com/yerdembek/AP2_assignment2/order-service/internal/repository"
	"github.com/yerdembek/AP2_assignment2/order-service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS environment variables")
	}

	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50052")
	paymentAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50051")
	dsn := buildDSN()

	// Connect to PostgreSQL
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := repository.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// gRPC client → Payment Service
	paymentClient, err := client.NewPaymentGRPCClient(paymentAddr)
	if err != nil {
		log.Fatalf("Failed to connect to payment service: %v", err)
	}
	defer paymentClient.Close()

	// Wire up Clean Architecture layers
	repo := repository.NewOrderRepository(db, dsn)
	uc := usecase.NewOrderUseCase(repo, paymentClient)

	// Start gRPC Streaming Server (Order Tracking)
	go startGRPCServer(grpcPort, uc)

	// Start REST HTTP Server
	r := gin.Default()
	handler := deliveryhttp.NewOrderHandler(uc)
	handler.RegisterRoutes(r)

	log.Printf("Order HTTP Server started on :%s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func startGRPCServer(port string, uc *usecase.OrderUseCase) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on :%s for gRPC: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	trackingSrv := deliverygrpc.NewOrderTrackingGRPCServer(uc)
	pb.RegisterOrderTrackingServiceServer(grpcServer, trackingSrv)

	log.Printf("Order gRPC Tracking Server started on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Order gRPC server error: %v", err)
	}
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "orders_db")
	ssl := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, name, ssl)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
