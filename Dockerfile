# syntax=docker/dockerfile:1

FROM golang:1.18.3-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY cmd ./cmd

RUN go build -o /usr/bin/chiefd ./cmd/chiefd/daemon.go

COPY config /etc/chiefd/

ENTRYPOINT [ "/usr/bin/chiefd" ]
