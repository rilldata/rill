<<<<<<< HEAD
REPO=rilldata
NAME=rill-developer
# PACKAGE_NAME          := github.com/rilldata/rill-developer
GOLANG_CROSS_VERSION  ?= v1.19.2

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

.PHONY: cli
cli: cli.prepare
	go build -o rill cli/main.go 

.PHONY: cli.prepare
cli.prepare:
	npm install
	npm run build
	rm -rf cli/pkg/web/embed/dist || true
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	rm -rf cli/pkg/examples/embed/dist || true
	mkdir -p cli/pkg/examples/embed/dist
	cp -r examples/ cli/pkg/examples/embed/dist/
=======
.PHONY: cli
cli:
	npm install
	npm run build
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	go build -o rill cli/main.go 
>>>>>>> c8a47306 (adding back the makefile for cli and proto)

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate
<<<<<<< HEAD
	npm run generate:runtime-client -w web-common
=======
>>>>>>> c8a47306 (adding back the makefile for cli and proto)
