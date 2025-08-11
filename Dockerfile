FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY mqtt-exporter-embedded /app/
RUN chmod +x /app/mqtt-exporter-embedded
COPY config.yaml /app/

EXPOSE 1883 8101 8080

CMD ["./mqtt-exporter-embedded", "--config", "config.yaml"]