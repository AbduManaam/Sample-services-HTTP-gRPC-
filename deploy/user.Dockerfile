# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/http-user ./cmd/http-user
COPY api ./api

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/http-user

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/user-service .

EXPOSE 9001

CMD ["./user-service"]
