FROM golang:1.25-alpine3.21

ENV GO111MODULE=on

WORKDIR /app/erupe

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

CMD [ "go", "run", "." ]