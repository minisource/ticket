# Build stage
FROM golang:1.26-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy go-common and go-sdk for local development
COPY ../go-common ../go-common
COPY ../go-sdk ../go-sdk

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ticket-service ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ticket-service .

# Copy locales
COPY --from=builder /app/locales ./locales

# Copy config if exists
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 5011

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5011/health || exit 1

# Run the application
CMD ["./ticket-service"]
