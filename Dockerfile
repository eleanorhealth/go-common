FROM golang:1.26.5-alpine

RUN apk upgrade --no-cache && \
    apk --no-cache add libc-dev gcc bash git

RUN apk update && \
    apk upgrade libssl3 libcrypto3 zlib

ENV CGO_ENABLED 1
# Let go download the toolchain go.mod declares (via the `go` directive) when it's
# newer than what's baked into this image, so the weekly go-version-bump PR doesn't
# also need a matching Dockerfile bump to keep building.
ENV GOTOOLCHAIN auto

WORKDIR /app
COPY . .

RUN go generate ./...
