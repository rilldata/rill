.PHONY: docs.generate
docs.generate:
	# Temporarily replaces ~/.rill/config.yaml to avoid including user-defined defaults in generated docs.
	# Sets version to the latest tag to simulate a production build, where certain commands are hidden.
	rm -rf docs/docs/reference/cli
	if [ -f ~/.rill/config.yaml ]; then mv ~/.rill/config.yaml ~/.rill/config.yaml.tmp; fi;
	go run -ldflags="-X main.Version=$(shell git describe --tags `git rev-list --tags --max-count=1`)" ./cli docs generate docs/docs/reference/cli/
	if [ -f ~/.rill/config.yaml.tmp ]; then mv ~/.rill/config.yaml.tmp ~/.rill/config.yaml; fi;

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
	git clone https://github.com/rilldata/rill-examples.git runtime/pkg/examples/embed/dist

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate --exclude-path rill/ui
	cd proto && buf generate --template buf.gen.openapi-admin.yaml --path rill/admin
	cd proto && buf generate --template buf.gen.openapi-runtime.yaml --path rill/runtime
	cd proto && buf generate --template buf.gen.ui.yaml
	npm run generate:runtime-client -w web-common
	npm run generate:client -w web-admin
