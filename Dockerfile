FROM golang:1.18-alpine

RUN apk --no-cache add libc-dev gcc git

ENV CGO_ENABLED 1

WORKDIR /app
COPY . .

RUN go generate ./...
