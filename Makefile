.PHONY: cli
cli:
	npm install
	npm run build
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	# go build -o rill cli/main.go
