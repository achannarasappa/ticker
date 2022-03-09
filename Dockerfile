FROM alpine:3

RUN apk add --no-cache tzdata

ENV TERM=xterm-256color

COPY ticker /ticker

VOLUME ["/.ticker.yaml"]

ENTRYPOINT ["/ticker"]
