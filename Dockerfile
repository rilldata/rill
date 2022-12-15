# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu:focal
RUN apt-get install -y ca-certificates
WORKDIR /project

COPY rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

ENTRYPOINT ["rill"]
CMD ["start"]
