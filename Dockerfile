# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu:focal

WORKDIR /app
COPY dist/linux_linux_amd64_v1/rill .

COPY scripts/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
