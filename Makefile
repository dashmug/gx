GO    := go
GOFMT := gofmt

.PHONY: all check test test-race vet fmt fmt-check lint bench cover tidy clean help

## all: fmt-check, vet, and test-race (default)
all: fmt-check vet test-race

## check: alias for all — use in CI
check: all

## test: run tests (no race detector)
test:
	$(GO) test ./...

## test-race: run tests with race detector (required gate)
test-race:
	$(GO) test -race ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## fmt: apply gofmt to all .go files in place
fmt:
	$(GOFMT) -w .

## fmt-check: exit non-zero if any .go files are unformatted
fmt-check:
	@test -z "$$($(GOFMT) -l .)" || { $(GOFMT) -l .; exit 1; }

## lint: run golangci-lint (requires golangci-lint to be installed)
lint:
	golangci-lint run

## bench: run benchmarks
bench:
	$(GO) test -bench=. -benchmem ./...

## cover: generate HTML coverage report → coverage.html
cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

## tidy: run go mod tidy
tidy:
	$(GO) mod tidy

## clean: remove generated coverage files
clean:
	rm -f coverage.out coverage.html

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## /  /'
