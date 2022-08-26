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
