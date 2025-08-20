# /home/ubuntu/jonobridge/pkg/xpot/Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum from the project root
COPY go.mod go.sum ./
RUN go mod download

# Copy the xpot source code from pkg/xpot
COPY ./ /app/xpot

WORKDIR /app/xpot

# Build the xpot binary
RUN CGO_ENABLED=0 GOOS=linux go build -o xpot main.go

# Create a minimal image with just the compiled binary
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /
# Copy only the binary from the builder stage
COPY --from=builder /app/xpot/xpot /xpot
# Expose the ports that the application listens on
EXPOSE 1883 

ENTRYPOINT ["/xpot", "-v"]
