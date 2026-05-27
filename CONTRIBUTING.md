# Contributing to tgs

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [golangci-lint](https://golangci-lint.run/welcome/install/)

## Build from source

```bash
git clone https://github.com/searchtgcli/tgs.git
cd tgs
make build
./tgs version
```

## Run tests

```bash
make test
```

## Run linter

```bash
make lint
```

## Pull requests

1. Fork the repository
2. Create a feature branch from `main`
3. Write tests for new functionality
4. Ensure `make lint` and `make test` pass
5. Submit a pull request

## Code style

Follow the project's golangci-lint configuration. Run `make lint` before submitting.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation
- `build:` build system or dependencies
- `ci:` CI/CD changes
- `test:` tests
- `refactor:` code refactoring
