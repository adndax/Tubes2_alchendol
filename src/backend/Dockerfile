FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o main .

# Check if data exists in builder stage
RUN ls -la /app/data/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=builder /app/main .

# Copy the data directory - make sure this is included
COPY --from=builder /app/data ./data

# Verify the data was copied
RUN ls -la /app/data/

EXPOSE 8080

CMD ["./main"]