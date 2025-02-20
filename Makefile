.PHONY: all
all: cli

.PHONE: cli-only
cli-only:
	cp scripts/install.sh cli/pkg/installscript/embed/install.sh
	go run scripts/embed_duckdb_ext/main.go
	go build -o rill cli/main.go

.PHONY: cli
cli: cli.prepare
	go build -o rill cli/main.go 

.PHONY: cli.prepare
cli.prepare:
	npm install
	npm run build
	rm -rf cli/pkg/web/embed/dist || true
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/* cli/pkg/web/embed/dist
	rm -rf runtime/pkg/examples/embed/dist || true
	mkdir -p runtime/pkg/examples/embed/dist
	git clone --quiet https://github.com/rilldata/rill-examples.git runtime/pkg/examples/embed/dist
	rm -rf runtime/pkg/examples/embed/dist/.git
	cp scripts/install.sh cli/pkg/installscript/embed/install.sh
	go run scripts/embed_duckdb_ext/main.go

.PHONY: coverage.go
coverage.go:
	rm -rf coverage/go.out
	mkdir -p coverage
	# Run tests with coverage output. First builds the list of packages to include in coverage, excluding generated code in 'proto/gen'.
	# NOTE(2024-03-01): Coverage fails on the generated code in 'proto/gen' without GOEXPERIMENT=nocoverageredesign. See https://github.com/golang/go/issues/55953.
	set -e ; \
		PACKAGES=$$(go list ./... | grep -v 'proto/gen/' | tr '\n' ',' | sed -e 's/,$$//' | sed -e 's/github.com\/rilldata\/rill/./g') ;\
		GOEXPERIMENT=nocoverageredesign go test ./... -short -v -coverprofile ./coverage/go.out -coverpkg $$PACKAGES
	go tool cover -func coverage/go.out

.PHONY: docs.generate
docs.generate:
	# Temporarily replaces ~/.rill/config.yaml to avoid including user-defined defaults in generated docs.
	# Sets version to the latest tag to simulate a production build, where certain commands are hidden.
	rm -rf docs/docs/reference/cli
	if [ -f ~/.rill/config.yaml ]; then mv ~/.rill/config.yaml ~/.rill/config.yaml.tmp; fi;
	go run -ldflags="-X main.Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)" ./cli docs generate docs/docs/reference/cli/
	if [ -f ~/.rill/config.yaml.tmp ]; then mv ~/.rill/config.yaml.tmp ~/.rill/config.yaml; fi;

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate --exclude-path rill/ui
	cd proto && buf generate --template buf.gen.openapi-admin.yaml --path rill/admin
	cd proto && buf generate --template buf.gen.openapi-runtime.yaml --path rill/runtime
	cd proto && buf generate --template buf.gen.local.yaml --path rill/local
	cd proto && buf generate --template buf.gen.ui.yaml
	npm run generate:runtime-client -w web-common
	npm run generate:client -w web-admin
