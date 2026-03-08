# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o portfolio-server ./cmd/portfolio

# Final Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/portfolio-server .

# Copy necessary directories
COPY --from=builder /app/data ./data
COPY --from=builder /app/tmpl ./tmpl
COPY --from=builder /app/static ./static
COPY --from=builder /app/assets ./assets

# Expose ports
EXPOSE 8080
EXPOSE 2222

# Ensure secure execution
USER nobody:nobody

# Run the binary
CMD ["./portfolio-server"]
