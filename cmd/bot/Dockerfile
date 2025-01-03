# Build
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bot ./cmd/bot/

# Runtime
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/bot .
COPY .env /app/.env
COPY locales /app/locales
CMD ["./bot"]
