Вот твой обновленный, профессиональный и "мега" README на английском языке, полностью адаптированный под твою структуру и пути, которые мы исправили.

Я сохранил стиль твоего примера, но наполнил его деталями твоего проекта.

---

# AP2 Assignment 2 — gRPC-Driven Microservices & Real-Time Streaming

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-Framework-4285F4?style=flat&logo=google)](https://grpc.io/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://www.docker.com/)

This project demonstrates a robust microservices architecture using **Go**, **gRPC (Unary & Streaming)**, and **Clean Architecture**. It features a reactive order processing system where status updates are pushed in real-time using **PostgreSQL LISTEN/NOTIFY**.

## 🔗 Repository Structure

| Module | Path | Purpose |
|---|---|---|
| **Proto Specs** | `/proto` | Source `.proto` definitions for Order and Payment services. |
| **Generated Code** | `/generated` | Auto-generated gRPC stubs used by all services. |
| **Order Service** | `/order-service` | Orchestrator service (REST + gRPC Client + gRPC Stream Server). |
| **Payment Service** | `/payment-service` | Payment processor (gRPC Server with Logging Interceptor). |
| **CLI Client** | `/client-cli` | Console tool to subscribe to real-time order updates. |

---

## 🏗 System Architecture

```
┌──────────────┐    REST (Gin)   ┌──────────────────────────────────────────────────┐
│   End User   │ ──────────────► │              Order Service                        │
│  (HTTP/REST) │                 │  :8080 (HTTP)  │  :50052 (gRPC Streaming)         │
└──────────────┘                 │                │                                  │
                                 │  ┌───────────┐ │                                  │
                                 │  │  Handler  │ │ OrderTrackingService             │
                                 │  └─────┬─────┘ │ SubscribeToOrderUpdates (stream) │
                                 │        │ UseCase│                                 │
                                 │  ┌─────▼─────┐ │                                  │
                                 │  │ Repository│ │                                  │
                                 │  └─────┬─────┘ │                                  │
                                 └────────┼────────┘                                 │
                                          │ gRPC (ProcessPayment)                    │
                                          ▼                                          │
                                 ┌──────────────────────────────────────────────────┐
                                 │             Payment Service  :50051               │
                                 │  ┌───────────────┐                               │
                                 │  │ gRPC Handler  │ ← Logging Interceptor (Bonus) │
                                 │  └──────┬────────┘                               │
                                 │         │ UseCase                                │
                                 │  ┌──────▼────────┐                               │
                                 │  │  Repository   │                               │
                                 │  └──────┬────────┘                               │
                                 └─────────┼────────────────────────────────────────┘
                                           │
                                 ┌─────────▼──────────┐    ┌────────────────────┐
                                 │  payments_db (PG)  │    │  orders_db (PG)    │
                                 └────────────────────┘    └────────────────────┘
                                                                     ▲
                                 Real-time streaming via PostgreSQL LISTEN/NOTIFY
```

### 🤖 Contract-First Workflow
The project uses a **Contract-First** approach. Definitions in `/proto` are compiled into `/generated` using a custom GitHub Actions workflow, ensuring type safety across all services.

---

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- Protoc (optional, for manual regeneration)

### Running via Docker Compose (Recommended)
This will spin up two PostgreSQL databases and both microservices automatically.
```bash
docker-compose up --build
```

**Services will be available at:**
- **Order Service (REST):** `http://localhost:8080`
- **Order Service (gRPC Stream):** `localhost:50052`
- **Payment Service (gRPC):** `localhost:50051`

---

## 🛠 API Reference

### Order Service (REST API)

| Method | Path | Description |
|--------|------|-------------|
| **POST** | `/orders` | Create an order (triggers gRPC call to Payment Service) |
| **GET**  | `/orders/:id` | Retrieve order details from DB |
| **PATCH**| `/orders/:id/status`| Manually update status (triggers DB NOTIFY → Stream) |

**Example: Create Order**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_007",
    "product_id": "premium_sub",
    "quantity": 1,
    "amount": 15.50,
    "currency": "USD"
  }'
```

---

## 📺 Real-Time Streaming Demo

To see the **gRPC Server Streaming** in action, use the built-in CLI client:

1. **Start the subscriber:**
   ```bash
   cd client-cli
   go run main.go -addr localhost:50052 -order <order_id_from_post_request>
   ```

2. **Trigger an update (in another terminal):**
   ```bash
   curl -X PATCH http://localhost:8080/orders/<order_id>/status \
     -H "Content-Type: application/json" \
     -d '{"status": "SHIPPED"}'
   ```

3. **Result:** The CLI client will instantly receive the update via **PostgreSQL LISTEN/NOTIFY** events without any polling!

---

## 💎 Features & Bonus Requirements

### 🛡 Clean Architecture
All services follow a strict 4-layer separation:
- `domain/`: Pure business entities.
- `usecase/`: Application business rules.
- `repository/`: Database implementations.
- `delivery/`: HTTP handlers and gRPC adapters.

### 📝 gRPC Logging Interceptor (Bonus)
The Payment Service implements a **UnaryServerInterceptor**. It captures and logs every incoming request, execution duration, and result:
```text
[gRPC] --> /payment.v1.PaymentService/ProcessPayment | request: {OrderId:ord_1... Amount:15.5}
[gRPC] <-- /payment.v1.PaymentService/ProcessPayment | duration: 1.5ms | OK
```

### ⚡ Manual Proto Regeneration
```bash
protoc --proto_path=proto \
  --go_out=generated --go_opt=paths=source_relative \
  --go-grpc_out=generated --go-grpc_opt=paths=source_relative \
  order/order.proto payment/payment.proto
```

---
👤 **Author:** [@yerdembek](https://github.com/yerdembek)  
📅 **Assignment:** AP2 Assignment #2 (Microservices & gRPC)
