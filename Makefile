DUCKDB_VERSION=0.4.0
LIB_PATH := $(shell pwd)/lib

ifeq ($(shell uname -s),Darwin)
LIB_EXT=dylib
ARCH_OS=osx-universal
LIBRARY_PATH := DYLD_LIBRARY_PATH=$(LIB_PATH)
else
LIB_EXT=so
ARCH_OS=linux-amd64
LIBRARY_PATH := LD_LIBRARY_PATH=$(LIB_PATH)
endif
LIBS := lib/libduckdb.$(LIB_EXT)
LDFLAGS := LIB=libduckdb.$(LIB_EXT) CGO_LDFLAGS="-L$(LIB_PATH)" $(LIBRARY_PATH) CGO_CFLAGS="-I$(LIB_PATH)"

$(LIBS):
	mkdir -p lib
	curl -Lo lib/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-$(ARCH_OS).zip
	cd lib; unzip -u libduckdb.zip

.PHONY: install
install: $(LIBS)
	$(LDFLAGS) go install -ldflags="-r $(LIB_PATH)" ./...

.PHONY: run
run: $(LIBS)
	$(LDFLAGS) go run -ldflags="-r $(LIB_PATH)" runtime/main.go stage.db

.PHONY: test
test: $(LIBS)
	$(LDFLAGS) go test -ldflags="-r $(LIB_PATH)" -v -race -count=1 ./...

.PHONY: clean
clean:
	rm -rf lib
