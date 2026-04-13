package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// Исправленные импорты:
	pb "github.com/yerdembek/AP2_assignment2/generated/payment"
	deliverygrpc "github.com/yerdembek/AP2_assignment2/payment-service/internal/delivery/grpc"
	"github.com/yerdembek/AP2_assignment2/payment-service/internal/interceptor"
	"github.com/yerdembek/AP2_assignment2/payment-service/internal/repository"
	"github.com/yerdembek/AP2_assignment2/payment-service/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS environment variables")
	}

	grpcPort := getEnv("GRPC_PORT", "50051")
	dsn := buildDSN()

	// Connect to database
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := repository.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")

	// Wire up dependencies (Clean Architecture)
	repo := repository.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	srv := deliverygrpc.NewPaymentGRPCServer(uc)

	// Start gRPC server with logging interceptor (Bonus)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryLoggingInterceptor),
	)
	pb.RegisterPaymentServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on :%s: %v", grpcPort, err)
	}

	log.Printf("Payment gRPC Server started on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	name := getEnv("DB_NAME", "payments_db")
	ssl := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, pass, name, ssl)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
