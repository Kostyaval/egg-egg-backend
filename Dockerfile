# see https://hub.docker.com/_/golang
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

COPY --chown=$USER:$USER ./ ./

RUN go mod download
RUN make build


FROM alpine:3.20 AS deploy
ARG USER="egg"
ENV USER=$USER
ENV UID=24680
ENV GID=24680

RUN apk update && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*

RUN addgroup -g "$GID" "$USER" && \
    adduser \
    --disabled-password \
    --gecos "" \
    --no-create-home \
    --ingroup "$USER" \
    --uid "$UID" \
    "$USER"

COPY --from=builder /home/$USER/server /bin/server

USER $USER
CMD ["/bin/server"]
