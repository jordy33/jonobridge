# /home/ubuntu/jonobridge/pkg/meitrackprotocol/Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the meitrackprotocol source code from pkg/meitrackprotocol
COPY ./ /app/meitrackprotocol

WORKDIR /app/meitrackprotocol

# Build the meitrackprotocol binary
RUN CGO_ENABLED=0 GOOS=linux go build -o meitrackprotocol main.go

# Create a minimal image with just the compiled binary
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /
# Copy only the binary from the builder stage
COPY --from=builder /app/meitrackprotocol/meitrackprotocol /meitrackprotocol
# Expose the ports that the application listens on
EXPOSE 1883 8080 80

ENTRYPOINT ["/meitrackprotocol", "-v"]
