PACKAGES       := $(shell find . -name "*.go" | grep -v -e vendor -e bindata | xargs -n1 dirname | sort -u)
TEST_FLAGS     := -race -count=1

default: test

.PHONY: lint
lint:
	golangci-lint run $(PACKAGES)

.PHONY: test
test:
	go test $(TEST_FLAGS) $(PACKAGES)

.PHONY: bench
bench:
	go test -bench=. -benchtime=10x
