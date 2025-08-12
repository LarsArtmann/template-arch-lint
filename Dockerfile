# Multi-stage Dockerfile for Go application
# Optimized for security, size, and production deployment

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Install templ for template generation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Generate templates
RUN templ generate

# Build the application
# CGO_ENABLED=1 for SQLite support
# Static binary for distroless compatibility
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server \
    ./cmd/server

# Runtime stage using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary
COPY --from=builder /build/server /server

# Copy web templates and static assets
COPY --from=builder /build/web /web

# Copy configuration files
COPY --from=builder /build/config.yaml /config.yaml
COPY --from=builder /build/config.production.yaml /config.production.yaml

# Create directory for SQLite database
USER 65532:65532

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/server", "-health-check"] || exit 1

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/server"]