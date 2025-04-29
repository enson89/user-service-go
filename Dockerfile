# Build Stage
FROM golang:1.24-alpine AS builder

# Install git and CA certificates for HTTPS
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags="-s -w"  \
    -o user-service cmd/api/main.go

# Final Stage
FROM scratch

# Copy CA certs for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/user-service /user-service

# Expose the service port
EXPOSE 8080

# Run as non-root user (UID 65532)
USER 65532:65532

# Entrypoint
ENTRYPOINT ["/user-service-go"]