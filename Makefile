.PHONY: cli
cli:
	npm install
	npm run build
	mkdir -p cli/pkg/web/embed/dist
	cp -r web-local/build/ cli/pkg/web/embed/dist
	go build -o rill cli/main.go 

.PHONY: proto.generate
proto.generate:
	cd proto && buf generate

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
        --env-file .release-env \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v `pwd`:/go/src/rilldata/rill-developer \
        -v `pwd`/sysroot:/sysroot \
        -v `pwd`/rill-binary-sa.json:/rill-binary-sa.json \
        -w /go/src/rilldata/rill-developer \
        goreleaser/goreleaser-cross:v1.19.2 \
		--rm-dist --skip-validate --skip-publish

.PHONY: release-cli
release-dry-run:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
        --env-file .release-env \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v `pwd`:/go/src/rilldata/rill-developer \
        -v `pwd`/sysroot:/sysroot \
        -v `pwd`/rill-binary-sa.json:/rill-binary-sa.json \
        -w /go/src/rilldata/rill-developer \
        goreleaser/goreleaser-cross:v1.19.2 \
        release --rm-dist --skip-validate
