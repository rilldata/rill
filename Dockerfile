# syntax = docker/dockerfile:1.1-experimental
FROM ubuntu 

RUN apt-get update && apt-get install -y curl unzip bash libdigest-sha-perl

RUN curl -s https://cdn.rilldata.com/install.sh | bash

ENTRYPOINT rill start
