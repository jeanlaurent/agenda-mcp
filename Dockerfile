# Build stage
FROM golang:1.24-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o agenda-mcp main.go

# Runtime stage
FROM scratch

# Copy ca-certificates from builder stage for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/agenda-mcp /agenda-mcp

# Expose port 8080 for OAuth callback
EXPOSE 8080

# Create a volume for credentials and token files
VOLUME ["/data"]

# Set working directory
WORKDIR /data

# Default command (can be overridden)
CMD ["/agenda-mcp", "mcp"] 