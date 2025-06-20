# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o gw2-mcp .

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/gw2-mcp /gw2-mcp

# Set the timezone
ENV TZ=UTC

# Expose port (though MCP typically uses stdio)
EXPOSE 8080

# Add labels
LABEL maintainer="AlyxPink"
LABEL description="Guild Wars 2 Model Context Provider Server"
LABEL version="1.0.0"

# Run the binary
ENTRYPOINT ["/gw2-mcp"]