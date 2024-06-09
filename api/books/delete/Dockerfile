# First stage: build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files from the root directory to /app in the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code from the delete directory to /app in the container
COPY api/books/delete/ .

# Build the application
RUN go build -o delete-service .

# Second stage: create the final, minimal image
FROM alpine:3.18

WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/delete-service .

# Expose port
EXPOSE 80

# Run the application
ENTRYPOINT ["./delete-service"]
