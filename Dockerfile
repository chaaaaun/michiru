FROM golang:1.24-alpine3.22 AS builder

RUN apk update
RUN apk add --no-cache musl-dev libxml2-dev gcc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ENV CGO_ENABLED=1
RUN go build -o /app/importer ./cmd/importer
RUN go build -o /app/deleter ./cmd/deleter

ENV CGO_ENABLED=0
RUN go build -o /app/server ./cmd/server

FROM alpine:3.22.0 AS importer

# Dependencies for the binary
RUN apk add --no-cache musl libxml2

# Use existing crond config to schedule the importer daily
COPY --from=builder /app/importer /etc/periodic/daily/

COPY --from=builder /app/deleter /root/

CMD ["/usr/sbin/crond", "-f", "-d", "0"]

FROM scratch AS server

COPY --from=builder /app/server /

CMD ["/server"]