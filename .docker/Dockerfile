# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary build tools and WebP libraries
RUN apk add --no-cache gcc musl-dev libwebp-dev

# Set the working directory
WORKDIR /app

# Copy the source code
COPY /src /app

# Build the application
ENV CGO_ENABLED=1
RUN go build -o esi-imageproxy

# Final stage
FROM alpine:latest

RUN apk add --no-cache libwebp

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/esi-imageproxy .

# Set the command to run the binary
CMD ["./esi-imageproxy"]
