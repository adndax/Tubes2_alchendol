FROM golang:1.24.2-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go mod download

RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Use exec form for ENTRYPOINT to avoid shell processing
ENTRYPOINT ["./main"]