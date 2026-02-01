# Build stage
FROM golang:1.25-alpine3.21 AS builder

ENV GO111MODULE=on

WORKDIR /app/erupe

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o erupe-ce .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app/erupe

COPY --from=builder /app/erupe/erupe-ce .

# Default command runs the compiled binary
CMD [ "./erupe-ce" ]
