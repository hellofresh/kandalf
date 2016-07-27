export GO15VENDOREXPERIMENT=1
GO_LINKER_FLAGS=-ldflags="-s -w"

APP_NAME=kandalf
MAIN_GO=$(CURDIR)/main.go

# Go vet settings
GO_VET_FILES=`go list -f '{{.Dir}}' ./... | grep -v /vendor/ | grep -v '$(APP_NAME)$$'`
GO_VET_FILES+=$(MAIN_GO)

# Useful directories
DIR_BUILD=$(CURDIR)/build
DIR_OUT=$(DIR_BUILD)/out
DIR_DEBIAN_TMP=$(DIR_OUT)/deb
DIR_RESOURCES=$(DIR_BUILD)/resources

# Remove the "v" prefix from version
VERSION=`$(DIR_OUT)/$(APP_NAME) -v | cut -d ' ' -f 3 | tr -d 'v'`

EXTERNAL_TOOLS=\
	github.com/Masterminds/glide

.build-linux:
	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 go build -o $(DIR_OUT)/linux/$(APP_NAME) $(GO_LINKER_FLAGS) $(MAIN_GO)

.build-osx:
	@echo Build OSX amd64
	env GOOS=darwin GOARCH=amd64 go build -o $(DIR_OUT)/darwin/$(APP_NAME) $(GO_LINKER_FLAGS) $(MAIN_GO)

# Default make target
build: .build-linux .build-osx

deb: .build-linux
	@echo Build debian package
	@mkdir $(DIR_DEBIAN_TMP)
	@mkdir -p $(DIR_DEBIAN_TMP)/etc/$(APP_NAME)
	@mkdir -p $(DIR_DEBIAN_TMP)/usr/local/bin
	@install -m 644 $(DIR_RESOURCES)/config.yml $(DIR_DEBIAN_TMP)/etc/$(APP_NAME)/config.yml
	@install -m 644 $(DIR_RESOURCES)/pipes.yml $(DIR_DEBIAN_TMP)/etc/$(APP_NAME)/pipes.yml
	@install -m 755 $(DIR_OUT)/linux/$(APP_NAME) $(DIR_DEBIAN_TMP)/usr/local/bin
	fpm -n $(APP_NAME) \
		-v $(VERSION) \
		-t deb \
		-s dir \
		-C $(DIR_DEBIAN_TMP) \
		-p $(DIR_OUT) \
		--config-files   /etc/$(APP_NAME) \
		--after-install  $(CURDIR)/build/debian/postinst \
		--after-remove   $(CURDIR)/build/debian/postrm \
		--deb-init       $(CURDIR)/build/debian/$(APP_NAME) \
		.
	@rm -rf $(DIR_DEBIAN_TMP)

docker-run:
	docker-compose up bridge

docker-up-env:
	docker-compose stop
	docker-compose rm --all --force
	docker-compose up -d elasticsearch
	docker-compose up -d kafka
	docker-compose up -d kibana
	docker-compose up -d logstash
	docker-compose up -d redis
	docker-compose up -d rmq
	sleep 2
	docker-compose exec rmq rabbitmqctl trace_on

fmt:
	@gofmt -s=true -w $(GO_VET_FILES)

run:
	go run -v $(MAIN_GO) -c=$(DIR_RESOURCES)/config.yml -p=$(DIR_RESOURCES)/pipes.yml

vet:
	@for vet_file in $(GO_VET_FILES); do \
		go tool vet $$vet_file; \
		if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Vet found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary before submitting the code for reviewal."; \
		fi \
	done

bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done
	@echo "Installing dependencies"; glide install

test:
	go test `go list ./... | grep -v /vendor/ | grep -v '$(APP_NAME)$$'`
