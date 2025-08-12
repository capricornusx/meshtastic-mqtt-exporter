FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY mqtt-exporter-embedded /app/
RUN chmod +x /app/mqtt-exporter-embedded
COPY config.yaml /app/

RUN addgroup -S nonroot \
    && adduser -S nonroot -G nonroot

USER nonroot

EXPOSE 1883 8100

CMD ["./mqtt-exporter-embedded", "--config", "config.yaml"]