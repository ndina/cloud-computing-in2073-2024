FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY web/ .

RUN go build -o web-service .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/web-service .

COPY --from=builder /app/views /app/views
COPY --from=builder /app/css /app/css

EXPOSE 80

ENTRYPOINT ["./web-service"]
