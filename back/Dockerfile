FROM golang:1.26.5-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /out \
    && CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/slotvio-api \
    ./cmd/api \
    && CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/slotvio-migrator \
    ./cmd/migrator

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S slotvio \
    && adduser -S -G slotvio slotvio

WORKDIR /app

COPY --from=builder --chown=slotvio:slotvio /out/slotvio-api ./slotvio-api
COPY --from=builder --chown=slotvio:slotvio /out/slotvio-migrator ./slotvio-migrator
COPY --chown=slotvio:slotvio config/config.yaml ./config/config.yaml
COPY --chown=slotvio:slotvio migrations/ ./migrations/

USER slotvio

EXPOSE 8080

ENTRYPOINT ["./slotvio-api"]
CMD ["--config=/app/config/config.yaml"]
