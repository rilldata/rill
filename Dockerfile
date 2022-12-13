# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu:focal

WORKDIR /app
COPY rill .

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
