# AP2 Assignment 2 — gRPC Migration & Contract-First Development

## Repository Links

| Repository | Purpose |
|---|---|
| **Proto Repo** | `github.com/RendersC/ap2-protos` — contains only `.proto` source files |
| **Generated Repo** | `github.com/RendersC/ap2-gen` — auto-generated `.pb.go` files (pushed by GitHub Actions) |

---

## Architecture

```
┌──────────────┐   REST (Gin)    ┌──────────────────────────────────────────────────┐
│   End User   │ ──────────────► │              Order Service                        │
│  (HTTP/REST) │                 │  :8080 (HTTP)  │  :50052 (gRPC Streaming)         │
└──────────────┘                 │                │                                   │
                                 │  ┌───────────┐ │                                   │
                                 │  │  Handler  │ │ OrderTrackingService              │
                                 │  └─────┬─────┘ │ SubscribeToOrderUpdates (stream)  │
                                 │        │ UseCase│                                   │
                                 │  ┌─────▼─────┐ │                                   │
                                 │  │ Repository│ │                                   │
                                 │  └─────┬─────┘ │                                   │
                                 └────────┼────────┘
                                          │ gRPC (ProcessPayment)
                                          ▼
                                 ┌──────────────────────────────────────────────────┐
                                 │             Payment Service  :50051               │
                                 │  ┌───────────────┐                               │
                                 │  │ gRPC Handler  │ ← Logging Interceptor (Bonus) │
                                 │  └──────┬────────┘                               │
                                 │         │ UseCase                                │
                                 │  ┌──────▼────────┐                               │
                                 │  │  Repository   │                               │
                                 │  └──────┬────────┘                               │
                                 └─────────┼──────────────────────────────────────-─┘
                                           │
                                 ┌─────────▼──────────┐    ┌────────────────────┐
                                 │  payments_db (PG)  │    │  orders_db (PG)    │
                                 └────────────────────┘    └────────────────────┘
                                                                     ▲
                                 Real-time streaming via PostgreSQL LISTEN/NOTIFY
```

### Contract-First Flow (GitHub Actions)

```
proto/ (Repo A)                           generated/ (Repo B)
├── payment/payment.proto  ──[push]──►  GitHub Actions  ──►  payment/payment.pb.go
└── order/order.proto                   (protoc-gen-go)      order/order.pb.go
                                                             └── go get github.com/RendersC/ap2-gen
```

---

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- `protoc` (only needed to regenerate protos manually)

---

## How to Run

### Option 1 — Docker Compose (Recommended)

```bash
docker compose up --build
```

This starts:
- `postgres-orders` on port 5432
- `postgres-payments` on port 5433
- `payment-service` gRPC on port 50051
- `order-service` HTTP on port 8080, gRPC streaming on port 50052

### Option 2 — Local (Two Terminals)

**Terminal 1 — Payment Service:**
```bash
cd payment-service
# Edit .env to point DB_HOST=localhost
go run ./cmd/main.go
```

**Terminal 2 — Order Service:**
```bash
cd order-service
# Edit .env to point DB_HOST=localhost, PAYMENT_SERVICE_ADDR=localhost:50051
go run ./cmd/main.go
```

---

## API Reference

### REST Endpoints (Order Service — port 8080)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orders` | Create order (calls Payment Service via gRPC internally) |
| GET  | `/orders/:id` | Get order by ID |
| PATCH | `/orders/:id/status` | Update order status (triggers DB notify → stream) |

**Create Order — Example:**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_42",
    "product_id": "prod_99",
    "quantity": 2,
    "amount": 49.99,
    "currency": "USD"
  }'
```

**Update Order Status (triggers streaming):**
```bash
curl -X PATCH http://localhost:8080/orders/<order_id>/status \
  -H "Content-Type: application/json" \
  -d '{"status": "SHIPPED"}'
```

---

## gRPC Streaming Demo

Subscribe to real-time order updates using the included CLI client:

```bash
cd client-cli
go run main.go -addr localhost:50052 -order <order_id>
```

Then in another terminal, update the order status:

```bash
curl -X PATCH http://localhost:8080/orders/<order_id>/status \
  -d '{"status": "PROCESSING"}' -H "Content-Type: application/json"
```

The CLI will immediately print the new status — no polling, powered by **PostgreSQL LISTEN/NOTIFY**.

---

## Regenerating Proto Code

```bash
cd proto/
protoc \
  --proto_path=. \
  --proto_path=<protoc-include-path>/include \
  --go_out=../generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=../generated \
  --go-grpc_opt=paths=source_relative \
  payment/payment.proto order/order.proto
```

Or push to the `proto` GitHub repository — GitHub Actions will generate and push to `ap2-generated` automatically.

---

## Clean Architecture Layers

```
internal/
├── domain/          # Pure domain entities — no framework dependencies
├── usecase/         # Business rules — depends only on domain & interfaces
├── repository/      # DB implementation of repository interface
└── delivery/
    ├── http/        # Gin REST handlers (Order Service only)
    └── grpc/        # gRPC handlers (server & client adapters)
```

The **Use Case layer is unchanged from Assignment 1** — only the delivery layer was updated to support gRPC.

---

## Bonus: gRPC Logging Interceptor

The Payment Service includes a **UnaryServerInterceptor** that logs every request:

```
[gRPC] --> /payment.v1.PaymentService/ProcessPayment | request: {OrderId:ord_123 Amount:49.99 ...}
[gRPC] <-- /payment.v1.PaymentService/ProcessPayment | duration: 2.3ms | OK
```

See `payment-service/internal/interceptor/logging.go`.
