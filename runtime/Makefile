RUNTIME_VERSION = $(shell cat version.txt | tr -d "[:space:]")

.PHONY: build
build:
	./scripts/setup_sql.sh -f
	go build -o main ./cmd

.PHONY: run
run:
	./scripts/setup_sql.sh
	go run ./cmd

.PHONY: test
test:
	./scripts/setup_sql.sh
	go test ./...

.PHONY: clean
clean:
	rm -rf sql/deps
	rm -f main
	rm -rf release

release/linux-amd64:
	mkdir -p release
	GOOS=linux GOARCH=amd64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		CC="zig cc -target x86_64-linux-musl" \
		CXX="zig c++ -target x86_64-linux-musl" \
		go build -ldflags "-X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -trimpath -o release/linux-amd64 ./cmd

release/windows-amd64.exe:
	mkdir -p release
	GOOS=windows GOARCH=amd64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
		CC="zig cc -target x86_64-windows-gnu" \
		CXX="zig c++ -target x86_64-windows-gnu" \
		go build -ldflags "-X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -trimpath -o release/windows-amd64.exe ./cmd

# NOTE: Move darwin builds to to Zig (for fully cross-platform builds) when the issues mentioned here have been solved: https://github.com/rilldata/rill-developer/pull/846

release/darwin-arm64:
	if [ "$(shell uname -s | tr '[:upper:]' '[:lower:]')" != "darwin" ]; then echo "Error: must run build on darwin"; false; fi
	mkdir -p release
	GOOS=darwin GOARCH=arm64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -ldflags "-X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -trimpath -o release/darwin-arm64 ./cmd

release/darwin-amd64:
	if [ "$(shell uname -s | tr '[:upper:]' '[:lower:]')" != "darwin" ]; then echo "Error: must run build on darwin"; false; fi
	mkdir -p release
	GOOS=darwin GOARCH=amd64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -ldflags "-X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -trimpath -o release/darwin-amd64 ./cmd
