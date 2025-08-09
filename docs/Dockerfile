FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o meshtastic-exporter-hook ./cmd/embedded-hook

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY --from=builder /app/meshtastic-exporter-hook .
COPY config.yaml .

EXPOSE 8100 1883

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8100/health || exit 1

CMD ["./meshtastic-exporter-hook", "--config", "config.yaml"]