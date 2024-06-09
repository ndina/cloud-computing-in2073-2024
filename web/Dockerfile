# First stage: build the Go application
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o web-service .

# Second stage: create the final, minimal image
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/web-service .
COPY --from=builder /app/views /app/views
COPY --from=builder /app/css /app/css

ENTRYPOINT ["./web-service"]
