# /home/ubuntu/jonobridge/pkg/huabaoprotocol/Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the huabaoprotocol source code from pkg/huabaoprotocol
COPY ./ /app/huabaoprotocol

WORKDIR /app/huabaoprotocol

# Build the huabaoprotocol binary
RUN CGO_ENABLED=0 GOOS=linux go build -o huabaoprotocol main.go

# Create a minimal image with just the compiled binary
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /
# Copy only the binary from the builder stage
COPY --from=builder /app/huabaoprotocol/huabaoprotocol /huabaoprotocol
# Expose the ports that the application listens on
EXPOSE 1883 8080 80

ENTRYPOINT ["/huabaoprotocol", "-v"]
