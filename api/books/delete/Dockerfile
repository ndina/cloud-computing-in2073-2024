FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY api/books/delete/ .

RUN go build -o delete-service .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/delete-service .

EXPOSE 80

ENTRYPOINT ["./delete-service"]
