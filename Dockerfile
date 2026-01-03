FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/worker /app/worker