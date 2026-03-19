.PHONY: build test lint install clean release release-dry-run tag

build:
	go build -o bin/rulekit ./

test:
	go test ./...

lint:
	golangci-lint run

install:
	go build -o $(shell go env GOPATH)/bin/rulekit ./

clean:
	rm -rf bin/ dist/

# Usage: make release VERSION=1.2.0
release: test
	@test -n "$(VERSION)" || (echo "Usage: make release VERSION=1.2.0"; exit 1)
	git tag v$(VERSION)
	git push origin v$(VERSION)

# Dry run — builds all platform binaries locally without publishing
release-dry-run:
	goreleaser release --snapshot --clean
