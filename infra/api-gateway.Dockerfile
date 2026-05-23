FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY services/api-gateway/ ./services/api-gateway
COPY shared/ ./shared

RUN go mod download

WORKDIR /app/services/api-gateway

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd/main.go

FROM alpine

COPY --from=builder /bin/api-gateway /bin/api-gateway

CMD ["/bin/api-gateway"]