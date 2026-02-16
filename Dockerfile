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
COPY --from=builder /build/www/ ./www/
COPY --from=builder /build/schemas/ ./schemas/
# bundled-schema/ is optional demo data, copy if present
RUN mkdir -p bundled-schema

# bin/ and savedata/ are mounted at runtime via docker-compose
# config.json is also mounted at runtime

USER erupe

ENTRYPOINT ["./erupe-ce"]
