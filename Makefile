VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  = -s -w \
	-X github.com/searchtgcli/tgs/internal/cmd.Version=$(VERSION) \
	-X github.com/searchtgcli/tgs/internal/cmd.Commit=$(COMMIT) \
	-X github.com/searchtgcli/tgs/internal/cmd.Date=$(DATE)

.PHONY: build lint test install clean

build:
	go build -ldflags '$(LDFLAGS)' -o tgs ./cmd/tgs/

lint:
	golangci-lint run ./...

test:
	go test ./... -v

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/tgs/

clean:
	rm -f tgs
