FROM alpine:3.13

ENV TERM=xterm-256color

RUN apk --no-cache add --update wget jq curl && rm -rf /var/cache/apk/* \
    && wget -c $(curl -s https://api.github.com/repos/achannarasappa/ticker/releases/latest | jq -r ".assets[] | select(.name | contains(\"linux-amd64\")) | .browser_download_url") \
    && tar -xf ticker*.tar.gz \
    && chmod +x ./ticker

VOLUME ["/.ticker.yaml"]

ENTRYPOINT  ["/ticker"]
