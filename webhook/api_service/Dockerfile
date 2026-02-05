# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy packages first
COPY queue /queue
COPY database /database

# Copy API source code
COPY api_service/ ./

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .
# Copy the .env file from builder stage
COPY --from=builder /app/.env .

COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]