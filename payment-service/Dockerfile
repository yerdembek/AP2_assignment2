FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY generated/ ./generated/

COPY payment-service/ ./payment-service/

WORKDIR /app/payment-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/payment-service/payment-service .

COPY payment-service/.env* ./

EXPOSE 50051

ENTRYPOINT ["./payment-service"]