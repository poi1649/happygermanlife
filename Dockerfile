FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Create a minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Create credentials directory
RUN mkdir -p /app/credentials

# Copy the binary from the builder stage
COPY --from=builder /app/server /app/server

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/server"]
