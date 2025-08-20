# /home/ubuntu/jonobridge/pkg/ruptelaprotocol/Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the ruptelaprotocol source code from pkg/ruptelaprotocol
COPY ./ /app/ruptelaprotocol

WORKDIR /app/ruptelaprotocol

# Build the ruptelaprotocol binary
RUN CGO_ENABLED=0 GOOS=linux go build -o ruptelaprotocol main.go

# Create a minimal image with just the compiled binary
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /
# Copy only the binary from the builder stage
COPY --from=builder /app/ruptelaprotocol/ruptelaprotocol /ruptelaprotocol
# Expose the ports that the application listens on
EXPOSE 1883 8080 80

ENTRYPOINT ["/ruptelaprotocol", "-v"]
