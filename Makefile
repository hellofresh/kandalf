NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: all lint test build

all: lint test build

build:
	@echo "$(OK_COLOR)==> Building... $(NO_COLOR)"
	/bin/sh -c "ARCH=$(ARCH) VERSION=${VERSION} ./build/build.sh"

test:
	@echo "$(OK_COLOR)==> Running tests$(NO_COLOR)"
	@CGO_ENABLED=0 go test -cover ./... -coverprofile=coverage.txt -covermode=atomic

lint:
	@echo "$(OK_COLOR)==> Linting with golangci-lint$(NO_COLOR)"
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v
