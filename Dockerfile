# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates wget

COPY rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

RUN groupadd -g 1000 rill \
    && useradd -m -u 1000 -s /bin/sh -g rill rill
USER rill

COPY scripts/extensions.sh .
RUN ./extensions.sh

ENTRYPOINT ["rill"]
CMD ["start"]
