# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y ca-certificates

COPY rill .
COPY web-admin/tests/setup/admin/.env .
RUN cp rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

ENTRYPOINT ["rill"]
CMD ["admin", "start"]
