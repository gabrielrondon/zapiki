# Multi-stage build for Zapiki

# Stage 1: Build
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zapiki-api cmd/api/main.go

# Build Worker binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zapiki-worker cmd/worker/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 zapiki && \
    adduser -D -u 1000 -G zapiki zapiki

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/zapiki-api .
COPY --from=builder /app/zapiki-worker .

# Copy start scripts
COPY --from=builder /app/start.sh .
COPY --from=builder /app/start-worker.sh .

# Copy scripts (for migrations if needed)
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/deployments/docker/schema.sql ./deployments/docker/

# Make start scripts executable
RUN chmod +x /app/start.sh /app/start-worker.sh

# Change ownership
RUN chown -R zapiki:zapiki /app

# Switch to app user
USER zapiki

# Expose port
EXPOSE 8080

# Use start script that auto-detects which service to run
CMD ["./start.sh"]
