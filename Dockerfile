# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o meals-app .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy config files
COPY --from=builder /build/config/config.yaml /app/config/
COPY --from=builder /build/config/config.production.yaml /app/config/

# Copy binary from builder
COPY --from=builder /build/meals-app .

# Set environment variable
ENV APP_ENV=production
# Make sure the server listens on all interfaces
ENV SERVER_HOST=0.0.0.0

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./meals-app"] 