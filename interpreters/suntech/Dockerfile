# /home/ubuntu/jonobridge/pkg/suntechprotocol/Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the suntechprotocol source code from pkg/suntechprotocol
COPY ./ /app/suntechprotocol

WORKDIR /app/suntechprotocol

# Build the suntechprotocol binary
RUN CGO_ENABLED=0 GOOS=linux go build -o suntechprotocol main.go

# Create a minimal image with just the compiled binary
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /
# Copy only the binary from the builder stage
COPY --from=builder /app/suntechprotocol/suntechprotocol /suntechprotocol
# Expose the ports that the application listens on
EXPOSE 1883 8080 80

ENTRYPOINT ["/suntechprotocol", "-v"]
