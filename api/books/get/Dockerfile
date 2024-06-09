FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY api/books/get/ .

RUN go build -o get-service .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/get-service .

EXPOSE 80

ENTRYPOINT ["./get-service"]
