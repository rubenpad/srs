# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy only the dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the migrations binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o ./run-migrations cmd/database/main.go

# Build the main application binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o ./srs cmd/api/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Create migrations directory
RUN mkdir -p /app/database/migrations

# Copy only the binaries from builder
COPY --from=builder /build/run-migrations /build/srs ./

# Copy migration files
COPY --from=builder /build/database/migrations/*.sql /app/database/migrations/

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["sh", "-c", "./run-migrations && ./srs"]
