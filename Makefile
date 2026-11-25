GO ?= go
COVERAGE_THRESHOLD ?= 90

.PHONY: all test race cover cover-html fmt fmt-check vet lint tidy check

all: check

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

cover:
	./scripts/coverage.sh $(COVERAGE_THRESHOLD)

cover-html: cover
	$(GO) tool cover -html=coverage.out -o coverage.html

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet:
	$(GO) vet ./...

lint:
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest ./...

tidy:
	$(GO) mod tidy

check: fmt-check vet test cover
