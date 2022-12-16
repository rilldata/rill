# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu:focal

RUN apt-get update && apt-get install -y ca-certificates

COPY rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

ENTRYPOINT ["rill"]
CMD ["start"]
