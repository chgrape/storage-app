FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY services/media-service/ ./services/media-service
COPY shared/ ./shared

RUN go mod download

WORKDIR /app/services/media-service

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/media-service ./cmd/api/main.go

FROM alpine

COPY --from=builder /bin/media-service /bin/media-service

CMD ["/bin/media-service"]