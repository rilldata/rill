FROM golang:1.20.5

RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir /rill-developer
WORKDIR /rill-developer
COPY go.mod .
COPY go.sum .
RUN go mod download

EXPOSE 8080/tcp
EXPOSE 9090/tcp


ENTRYPOINT ["go"]
CMD ["run", "/rill-developer/cli/main.go", "admin", "start"]