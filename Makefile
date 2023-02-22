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
	rm -rf cli/pkg/examples/embed/dist || true
	mkdir -p cli/pkg/examples/embed/dist
	cp -r examples/* cli/pkg/examples/embed/dist/

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate --exclude-path rill/ui
	cd proto && buf generate --template buf.gen.openapi-admin.yaml --path rill/admin
	cd proto && buf generate --template buf.gen.openapi-runtime.yaml --path rill/runtime
	cd proto && buf generate --template buf.gen.ui.yaml
	npm run generate:runtime-client -w web-common
