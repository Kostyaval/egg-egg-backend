# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine3.20 AS builder

ARG USER="egg"
ARG UID=24680
ARG GID=24680

RUN apk update && \
    apk add --no-cache make git && \
    addgroup -g $GID $USER && \
    adduser -D -u $UID -G $USER $USER && \
    rm -rf /var/cache/apk/*

USER $USER
WORKDIR /home/$USER

# Copy and download dependencies separately to leverage caching
COPY --chown=$USER:$USER go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY --chown=$USER:$USER ./ ./

# Install golangci-lint, run lint and tests
RUN apk add --no-cache golangci-lint --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community && \
    make lint && \
    make test

# Build the application
RUN make build

# Final stage
FROM alpine:3.20 AS deploy

ARG USER="egg"
ENV USER=$USER
ENV UID=24680
ENV GID=24680

RUN apk update && \
    apk add --no-cache ca-certificates && \
    addgroup -g $GID $USER && \
    adduser -D -u $UID -G $USER $USER && \
    rm -rf /var/cache/apk/*

COPY --from=builder /home/$USER/server /bin/server
COPY --from=builder /home/$USER/rules.yml /etc/egg/rules.yml

USER $USER
CMD ["/bin/server"]