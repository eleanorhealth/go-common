FROM golang:1.19-alpine

RUN apk --no-cache add libc-dev gcc git

RUN apk update && \
    apk upgrade libssl3 libcrypto3

ENV CGO_ENABLED 1

WORKDIR /app
COPY . .

RUN go generate ./...
