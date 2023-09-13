default: test

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: bench
bench:
	go test -bench=. -benchtime=10x
