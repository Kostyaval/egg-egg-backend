# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine3.20 AS builder

ARG USER="egg"
ARG UID=24680
ARG GID=24680

RUN apk update && \
    apk add --no-cache make git && \
    rm -rf /var/cache/apk/*

RUN addgroup -g "$GID" "$USER" && \
    adduser \
    --disabled-password \
    --gecos "" \
    --ingroup "$USER" \
    --uid "$UID" \
    "$USER"

USER $USER
WORKDIR /home/$USER

# Copy and download dependencies separately to leverage caching
COPY --chown=$USER:$USER go.mod go.sum ./
RUN go mod download

# Test and lint stage
FROM builder AS tester
COPY --chown=$USER:$USER ./ ./
RUN apk add --no-cache golangci-lint --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community && \
    make lint && \
    make test

# Build stage
FROM builder AS compiler
COPY --from=tester /home/$USER ./
RUN make build

# Final stage
FROM alpine:3.20 AS deploy
ARG USER="egg"
ENV USER=$USER
ENV UID=24680
ENV GID=24680

RUN apk update && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/* && \
    addgroup -g "$GID" "$USER" && \
    adduser \
    --disabled-password \
    --gecos "" \
    --no-create-home \
    --ingroup "$USER" \
    --uid "$UID" \
    "$USER"

COPY --from=compiler /home/$USER/server /bin/server
COPY --from=compiler /home/$USER/rules.yml /etc/egg/rules.yml
USER $USER
CMD ["/bin/server"]