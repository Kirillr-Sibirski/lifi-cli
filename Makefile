BINARY ?= bin/lifi
MODULE ?= github.com/Kirillr-Sibirski/lifi-cli/internal/cli.version
VERSION ?= dev

.PHONY: build test vet snapshot smoke-read install-head

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X $(MODULE)=$(VERSION)" -o $(BINARY) ./cmd/lifi

test:
	go test ./...

vet:
	go vet ./...

snapshot:
	goreleaser release --snapshot --clean

smoke-read:
	LIFI_SMOKE=1 go test ./internal/cli -run TestLiveSmokeReadPath -count=1

install-head:
	brew install --HEAD Kirillr-Sibirski/lifi-cli/lifi
