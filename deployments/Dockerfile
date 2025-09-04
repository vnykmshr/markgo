# Multi-stage Dockerfile for markgo

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o markgo ./cmd/server

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 markgo && \
    adduser -D -u 1001 -G markgo markgo

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/markgo .

# Copy static assets and templates
COPY --chown=markgo:markgo web/ ./web/
COPY --chown=markgo:markgo articles/ ./articles/

# Create necessary directories
RUN mkdir -p logs && chown markgo:markgo logs

# Switch to non-root user
USER markgo

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Set default environment variables
ENV ENVIRONMENT=production
ENV PORT=3000
ENV ARTICLES_PATH=./articles
ENV STATIC_PATH=./web/static
ENV TEMPLATES_PATH=./web/templates

# Run the application
CMD ["./markgo"]
