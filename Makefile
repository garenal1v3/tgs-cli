VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
TGS_API_ID   ?= 0
TGS_API_HASH ?= ""
LDFLAGS  = -s -w \
	-X github.com/searchtgcli/tgs/internal/cmd.Version=$(VERSION) \
	-X github.com/searchtgcli/tgs/internal/cmd.Commit=$(COMMIT) \
	-X github.com/searchtgcli/tgs/internal/cmd.Date=$(DATE) \
	-X github.com/searchtgcli/tgs/internal/telegram.defaultAPIID=$(TGS_API_ID) \
	-X github.com/searchtgcli/tgs/internal/telegram.defaultAPIHash=$(TGS_API_HASH)

GOLANGCI_LINT_VERSION ?= v2.12.2

.PHONY: build lint lint-docker fmt test install clean

build:
	go build -ldflags '$(LDFLAGS)' -o tgs ./cmd/tgs/

lint:
	golangci-lint run ./...

lint-docker:
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

test:
	go test ./... -v

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/tgs/

clean:
	rm -f tgs
