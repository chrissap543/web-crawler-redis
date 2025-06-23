# Multi-stage build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app
RUN mkdir -p /app/data

COPY --from=builder /app/scraper .

RUN chmod +x ./scraper
CMD ["./scraper"]
