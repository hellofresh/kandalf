NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: clean build

VERSION ?= "0.0.0-dev-$(shell git rev-parse --short HEAD)"

build:
	@echo "$(OK_COLOR)==> Building (v${VERSION}) ... $(NO_COLOR)"
	CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-X main.version=${VERSION}" -o "kandalf"

test:
	@echo "$(OK_COLOR)==> Running tests$(NO_COLOR)"
	@go test -cover -coverprofile=coverage.txt -covermode=atomic ./...

lint:
	@echo "$(OK_COLOR)==> Linting with golangci-lint running in docker container$(NO_COLOR)"
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v

clean:
	@echo "$(OK_COLOR)==> Cleaning project$(NO_COLOR)"
	@go clean
