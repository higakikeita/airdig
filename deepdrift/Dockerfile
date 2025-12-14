# UI build stage
FROM node:20-alpine AS ui-builder

WORKDIR /app/ui

# Copy UI package files
COPY ui/package*.json ./
RUN npm ci

# Copy UI source and build
COPY ui/ ./
RUN npm run build

# Go build stage
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy UI build from ui-builder
COPY --from=ui-builder /app/ui/build ./ui/build

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o deepdrift ./cmd/deepdrift

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from go-builder
COPY --from=go-builder /app/deepdrift .

# Copy UI build
COPY --from=go-builder /app/ui/build ./ui/build

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./deepdrift"]
CMD ["--command", "server", "--host", "0.0.0.0", "--port", "8080"]
