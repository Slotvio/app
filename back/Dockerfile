FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/slotvio-api \
    ./cmd/api


FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S slotvio \
    && adduser -S -G slotvio slotvio

WORKDIR /app

COPY --from=builder --chown=slotvio:slotvio /app/slotvio-api ./slotvio-api
COPY --chown=slotvio:slotvio config/config.yaml ./config/config.yaml

USER slotvio

EXPOSE 8080

ENTRYPOINT ["./slotvio-api"]
CMD ["--config=/app/config/config.yaml"]