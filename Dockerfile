# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates

COPY rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

RUN groupadd -g 1001 rill \
    && useradd -m -u 1001 -s /bin/sh -g rill rill
USER rill

RUN rill runtime install-duckdb-extensions

ENTRYPOINT ["rill"]
CMD ["start"]
