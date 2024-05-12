# Use the official Golang Alpine image with Go version 1.22
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the entire project into the container's working directory
COPY . .

# Download and cache Go modules
RUN go mod tidy

# Build the executable from main.go
RUN go build -o myapp ./cmd

# Command to run the executable
CMD ["./myapp"]
