FROM golang:1.22.0

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# copy code files
COPY cli cli
COPY proto proto
COPY runtime runtime
COPY admin admin

# copy ui files
RUN mkdir -p cli/pkg/web/embed/dist
COPY web-local/build/* cli/pkg/web/embed/dist

# build the binary, using a dummy version to enable admin and runtime commands
RUN go build -o rill cli/main.go -ldflags="-s -w -X main.Version=v0.0.1"

RUN cp rill /usr/local/bin
RUN chmod 777 /usr/local/bin/rill

RUN mkdir project
RUN touch project/rill.yaml

ENTRYPOINT ["rill"]
CMD ["start", "project"]
