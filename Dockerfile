# Build stage
FROM golang:1.25-alpine3.21 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o erupe-ce .

# Runtime stage
FROM alpine:3.21

RUN adduser -D -h /app erupe
WORKDIR /app

COPY --from=builder /build/erupe-ce .

# docker-compose mounts docker/bin/ and docker/savedata/ to /app/bin and
# /app/savedata respectively; config.json is also mounted at runtime

USER erupe

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./erupe-ce"]
