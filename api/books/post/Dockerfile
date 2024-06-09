FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY api/books/post/ .

RUN go build -o post-service .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/post-service .

EXPOSE 80

ENTRYPOINT ["./post-service"]
