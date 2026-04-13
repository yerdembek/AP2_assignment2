package grpc

import (
	"log"
	"time"

	// Исправлен путь к сгенерированному коду заказов
	pb "github.com/yerdembek/AP2_assignment2/generated/order"
	// Исправлен путь к внутреннему пакету usecase
	"github.com/yerdembek/AP2_assignment2/order-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderTrackingGRPCServer implements the generated OrderTrackingServiceServer.
// It streams real-time order status updates backed by PostgreSQL LISTEN/NOTIFY.
type OrderTrackingGRPCServer struct {
	pb.UnimplementedOrderTrackingServiceServer
	uc *usecase.OrderUseCase
}

func NewOrderTrackingGRPCServer(uc *usecase.OrderUseCase) *OrderTrackingGRPCServer {
	return &OrderTrackingGRPCServer{uc: uc}
}

// SubscribeToOrderUpdates streams order status changes to the client.
// Each time the order status changes in the DB (via pg_notify), the update is
// immediately pushed over the stream.
func (s *OrderTrackingGRPCServer) SubscribeToOrderUpdates(
	req *pb.OrderRequest,
	stream pb.OrderTrackingService_SubscribeToOrderUpdatesServer,
) error {
	if req.OrderId == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	ctx := stream.Context()
	log.Printf("SubscribeToOrderUpdates: starting stream for order_id=%s", req.OrderId)

	updatesCh, err := s.uc.SubscribeToOrderUpdates(ctx, req.OrderId)
	if err != nil {
		return status.Errorf(codes.NotFound, "cannot subscribe to order %s: %v", req.OrderId, err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("SubscribeToOrderUpdates: client disconnected for order_id=%s", req.OrderId)
			return nil
		case newStatus, ok := <-updatesCh:
			if !ok {
				// Channel closed — no more updates.
				return nil
			}
			update := &pb.OrderStatusUpdate{
				OrderId:   req.OrderId,
				Status:    newStatus,
				UpdatedAt: timestamppb.New(time.Now().UTC()),
				Message:   "Order status updated to: " + newStatus,
			}
			if err := stream.Send(update); err != nil {
				log.Printf("SubscribeToOrderUpdates: send error for order_id=%s: %v", req.OrderId, err)
				return status.Errorf(codes.Internal, "stream send error: %v", err)
			}
			log.Printf("SubscribeToOrderUpdates: sent update order_id=%s status=%s", req.OrderId, newStatus)
		}
	}
}
