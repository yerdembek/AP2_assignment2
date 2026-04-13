FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY generated/ ./generated/

COPY order-service/ ./order-service/

WORKDIR /app/order-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/order-service/order-service .

COPY order-service/.env* ./

EXPOSE 8080 50052

ENTRYPOINT ["./order-service"]