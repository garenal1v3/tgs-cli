VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w \
	-X github.com/garenal1v3/tgs-cli/internal/cmd.Version=$(VERSION) \
	-X github.com/garenal1v3/tgs-cli/internal/cmd.Commit=$(COMMIT) \
	-X github.com/garenal1v3/tgs-cli/internal/cmd.Date=$(DATE)

GOLANGCI_LINT_VERSION ?= v2.12.2

.PHONY: build lint lint-docker fmt test install clean docs ci hooks

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

docs:
	@command -v hugo >/dev/null || { echo "hugo not found: brew install hugo"; exit 1; }
	cd docs && npm install postcss postcss-cli autoprefixer 2>/dev/null; hugo server

ci:
	docker build -f Dockerfile.ci -t tgs-ci:local .
	docker run --rm tgs-ci:local golangci-lint run ./...
	docker run --rm tgs-ci:local make test
	docker run --rm tgs-ci:local sh -c "cd docs && hugo --minify"

hooks:
	git config core.hooksPath .githooks

clean:
	rm -f tgs
