FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet ./internal/app/

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/wallet .
COPY config.env .
EXPOSE 8080
CMD ["./wallet"]
