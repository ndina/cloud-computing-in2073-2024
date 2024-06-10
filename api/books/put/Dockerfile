FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ../../go.mod ../../go.sum ./

RUN go mod download

COPY api/books/put/ .

RUN go build -o put-service .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/put-service .

EXPOSE 80

ENTRYPOINT ["./put-service"]
