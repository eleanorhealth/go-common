FROM golang:1.26-alpine

RUN apk upgrade --no-cache && \
    apk --no-cache add libc-dev gcc bash git

RUN apk update && \
    apk upgrade libssl3 libcrypto3 zlib

ENV CGO_ENABLED 1

WORKDIR /app
COPY . .

RUN go generate ./...
