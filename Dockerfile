# Start from the latest golang base image
FROM golang:1.21.3 AS builder

# Add Maintainer Info
LABEL maintainer="JYY <yourrubber@duck.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download && go mod vendor

# Copy the source from the current directory to the Working Directory inside the container
COPY . .
COPY .env /root/.env

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o main .

# Expose port 50051 to the outside world
EXPOSE 50051

# Command to run the executable
CMD ["./main"]