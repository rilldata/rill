FROM ubuntu

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y ca-certificates

COPY rill .
COPY web-admin/tests/setup/admin/.env .
RUN cp rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

RUN rill runtime install-duckdb-extensions

ENTRYPOINT ["rill"]
CMD ["runtime", "start"]
