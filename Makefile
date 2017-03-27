GO_LINKER_FLAGS=-ldflags="-s -w"
CGO_ENABLED=0

APP_NAME=kandalf
APP_SRC=$(CURDIR)/cmd/kandalf/main.go

GO_PROJECT_FILES=`go list -f '{{.Dir}}' ./... | grep -v /vendor/`
GO_PROJECT_PACKAGES=`go list ./... | grep -v /vendor/`

# Useful directories
DIR_OUT=$(CURDIR)/out
DIR_OUT_LINUX=$(DIR_OUT)/linux
DIR_RESOURCES=$(CURDIR)/ci/resources
DOCKER_COMPOSE_FILE=$(CURDIR)/docker-compose.test.yml

EXTERNAL_TOOLS=\
	github.com/kisielk/errcheck \
	github.com/Masterminds/glide

.vet:
	@echo "Checking for suspicious constructs"
	@for project_file in $(GO_PROJECT_FILES); do \
		go tool vet $$project_file; \
		if [ $$? -ne 0 ]; then \
			echo ""; \
			echo "Vet found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary."; \
			exit 1;\
		fi \
	done

.errcheck:
	@echo "Checking the go files for unchecked errors"
	@for project_file in $(GO_PROJECT_FILES); do \
		if [ -f $$project_file ]; then \
			errcheck $$project_file; \
		else \
			errcheck $$(find $$project_file -type f); \
		fi; \
		if [ $$? -ne 0 ]; then \
			echo ""; \
			echo "Found not handled returning errors. Please check them and fix if necessary."; \
			exit 1;\
		fi \
	done

# Default make target
build: build-linux build-osx

build-linux:
	@echo Build Linux amd64
	@env GOOS=linux GOARCH=amd64 go build -o $(DIR_OUT_LINUX)/$(APP_NAME) $(GO_LINKER_FLAGS) $(APP_SRC)

build-osx:
	@echo Build OSX amd64
	@env GOOS=darwin GOARCH=amd64 go build -o $(DIR_OUT)/darwin/$(APP_NAME) $(GO_LINKER_FLAGS) $(APP_SRC)

# Launch all checks
check: .vet .errcheck

# Run the application in docker (only for testing purposes)
docker-run:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up bridge

# Bootstrap and up docker environment (only for testing purposes)
docker-up-env:
	docker-compose -f $(DOCKER_COMPOSE_FILE) stop
	docker-compose -f $(DOCKER_COMPOSE_FILE) rm --force
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d kafka
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d redis
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d rmq
	sleep 4
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec rmq rabbitmqctl trace_on

# Format the source code
fmt:
	@gofmt -s=true -w $(GO_PROJECT_FILES)

# Run the program from CLI without compilation for testing purposes
run:
	go run $(APP_SRC) -c=$(DIR_RESOURCES)/config.yml -p=$(DIR_RESOURCES)/pipes.yml

# Bootstrap vendoring tool and dependencies
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done
	@echo "Installing dependencies"; glide install

# Launch tests
test:
	@go test $(GO_PROJECT_PACKAGES)
