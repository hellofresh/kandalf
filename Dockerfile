FROM alpine:latest
WORKDIR /app
ENTRYPOINT ./kandalf -c /config/config.yml -p /config/pipes.yml
STOPSIGNAL SIGINT
