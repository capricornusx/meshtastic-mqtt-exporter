FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY meshtastic-exporter-embedded .
COPY config.yaml .

EXPOSE 8100 1883

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8100/health || exit 1

CMD ["./meshtastic-exporter-embedded", "--config", "config.yaml"]