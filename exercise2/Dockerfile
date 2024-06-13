# Use the official Golang image as the builder stage
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod file
COPY go.mod ./

# Download all dependencies. This will generate the go.sum file
RUN go mod tidy

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o shred-tool .

# Run tests
RUN go test -v ./...

RUN echo "Build complete."

# Run tests
RUN go test -v  -coverpkg=shred-tool/shred -coverprofile=coverage.out ./...
RUN go tool cover -func=coverage.out

RUN echo "All tests passed successfully."

ENTRYPOINT ["./shred-tool"]
