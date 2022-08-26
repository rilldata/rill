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

release/darwin-arm64: $(MACOS_SDK)
	mkdir -p release
	GOOS=darwin GOARCH=arm64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		CGO_LDFLAGS="-mmacosx-version-min=${MACOS_MIN_VERSION} --sysroot ${MACOS_SDK} -F/System/Library/Frameworks -L/usr/lib" \
		CC="zig cc -mmacosx-version-min=${MACOS_MIN_VERSION} -target aarch64-macos-gnu -isysroot ${MACOS_SDK} -iwithsysroot /usr/include -iframeworkwithsysroot /System/Library/Frameworks" \
		CXX="zig c++ -mmacosx-version-min=${MACOS_MIN_VERSION} -target aarch64-macos-gnu -isysroot ${MACOS_SDK} -iwithsysroot /usr/include -iframeworkwithsysroot /System/Library/Frameworks" \
		go build -ldflags "-s -w -X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -buildmode=pie -trimpath -o release/darwin-arm64 ./cmd

release/darwin-amd64: $(MACOS_SDK)
	mkdir -p release
	GOOS=darwin GOARCH=amd64 ./scripts/setup_sql.sh -f
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		CGO_LDFLAGS="-mmacosx-version-min=${MACOS_MIN_VERSION} --sysroot ${MACOS_SDK} -F/System/Library/Frameworks -L/usr/lib" \
		CC="zig cc -mmacosx-version-min=${MACOS_MIN_VERSION} -target x86_64-macos-gnu -isysroot ${MACOS_SDK} -iwithsysroot /usr/include -iframeworkwithsysroot /System/Library/Frameworks" \
		CXX="zig c++ -mmacosx-version-min=${MACOS_MIN_VERSION} -target x86_64-macos-gnu -isysroot ${MACOS_SDK} -iwithsysroot /usr/include -iframeworkwithsysroot /System/Library/Frameworks" \
		go build -ldflags "-s -w -X 'github.com/rilldata/rill/runtime.Version=${RUNTIME_VERSION}'" -buildmode=pie -trimpath -o release/darwin-amd64 ./cmd

##
# MACOS SDK HACK
##

MACOS_MIN_VERSION=11.3

OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
ifeq ($(OS),darwin)
  MACOS_SDK = $(shell xcrun --show-sdk-path)
else
  MACOS_SDK = $(shell pwd)/release/tmp/MacOSX11.3.sdk
endif

$(MACOS_SDK):
	mkdir -p $(MACOS_SDK)
	curl -L "https://github.com/phracker/MacOSX-SDKs/releases/download/11.3/MacOSX11.3.sdk.tar.xz" | tar -J -x -C "$(shell pwd)/release/tmp"
