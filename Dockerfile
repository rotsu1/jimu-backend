# STAGE 1: Build the binary
FROM golang:1.25-alpine AS builder

# Install build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy and download dependencies (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app into a static binary named 'main'
# CGO_ENABLED=0 makes the binary independent of system libraries
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# STAGE 2: Create the final tiny image
FROM alpine:latest

# Install CA certificates (required for connecting to Supabase via HTTPS/SSL)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your app runs on
EXPOSE 8080

# Run the app
CMD ["./main"]