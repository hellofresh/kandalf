export GOPATH=$(CURDIR)/.go

APP_NAME=kandalf
OUTDIR=$(CURDIR)/out
# Remove the "v" prefix from version
VERSION=`$(OUTDIR)/$(APP_NAME) -v | cut -d ' ' -f 3 | tr -d 'v'`
DEBIAN_TMP=$(OUTDIR)/deb

$(OUTDIR)/$(APP_NAME): $(CURDIR)/src/main.go
	go build -o $(OUTDIR)/$(APP_NAME) -ldflags="-s -w" $(CURDIR)/src/main.go

deb: $(OUTDIR)/$(APP_NAME)
	mkdir $(DEBIAN_TMP)
	mkdir -p $(DEBIAN_TMP)/etc/$(APP_NAME)
	mkdir -p $(DEBIAN_TMP)/usr/local/bin
	install -m 644 $(CURDIR)/data/config.yml $(DEBIAN_TMP)/etc/$(APP_NAME)/config.yml
	install -m 644 $(CURDIR)/data/pipes.yml $(DEBIAN_TMP)/etc/$(APP_NAME)/pipes.yml
	install -m 755 $(OUTDIR)/$(APP_NAME) $(DEBIAN_TMP)/usr/local/bin
	fpm -n $(APP_NAME) \
		-v $(VERSION) \
		-t deb \
		-s dir \
		-C $(DEBIAN_TMP) \
		-p $(OUTDIR) \
		--config-files   /etc/$(APP_NAME) \
		--after-install  $(CURDIR)/debian/postinst \
		--after-remove   $(CURDIR)/debian/postrm \
		--deb-init       $(CURDIR)/debian/$(APP_NAME) \
		.
	rm -rf $(DEBIAN_TMP)

dep-install:
	go get github.com/bshuster-repo/logrus-logstash-hook
	go get github.com/olebedev/config
	go get github.com/Sirupsen/logrus
	go get github.com/streadway/amqp
	go get github.com/urfave/cli
	go get gopkg.in/redis.v3
	go get gopkg.in/Shopify/sarama.v1
	go get gopkg.in/yaml.v2

docker-build:
	docker run --rm -it \
		-v $(GOPATH):/gopath \
		-v $(OUTDIR):/out \
		-v $(CURDIR)/src:/app \
		-e "GOPATH=/gopath" \
		-w /app golang:latest \
		sh -c 'go build -o /out/$(APP_NAME)-linux-amd64 -ldflags="-s -w" main.go'

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
	gofmt -s=true -w $(CURDIR)/src

run:
	go run -v $(CURDIR)/src/main.go -c=$(CURDIR)/data/config.yml -p=$(CURDIR)/data/pipes.yml

test:
	go test ./...
