FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY ./bin/app /usr/bin/app

CMD ["/usr/bin/app"]