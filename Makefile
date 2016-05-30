export GOPATH=$(CURDIR)/.go

APP_NAME = kandalf
OUTDIR=$(CURDIR)/out
VERSION=`$(OUTDIR)/$(APP_NAME) -v | cut -d ' ' -f 3`

$(OUTDIR)/$(APP_NAME): $(CURDIR)/src/main.go
	go build -o $(OUTDIR)/$(APP_NAME) -ldflags="-s -w" $(CURDIR)/src/main.go

dep-install:
	go get github.com/bshuster-repo/logrus-logstash-hook
	go get github.com/olebedev/config
	go get github.com/Sirupsen/logrus
	go get github.com/urfave/cli
	go get gopkg.in/yaml.v2

dep-update:
	go get -u github.com/bshuster-repo/logrus-logstash-hook
	go get -u github.com/olebedev/config
	go get -u github.com/Sirupsen/logrus
	go get -u github.com/urfave/cli
	go get -u gopkg.in/yaml.v2

docker-build:
	docker run --rm -it \
		-v $(GOPATH):/gopath \
		-v $(OUTDIR):/out \
		-v $(CURDIR)/src:/app \
		-e "GOPATH=/gopath" \
		-w /app golang:alpine \
		sh -c 'go build -o /out/$(APP_NAME) -ldflags="-s -w" main.go'

docker-run:
	docker-compose up bridge

docker-up-env:
	docker-compose stop
	docker-compose rm --all --force
	docker-compose up -d kafka
	docker-compose up -d redis
	docker-compose up -d rmq
	sleep 2
	docker-compose exec rmq rabbitmqctl trace_on

fmt:
	gofmt -s=true -w $(CURDIR)/src

run:
	go run -v $(CURDIR)/src/main.go -c=$(CURDIR)/data/config.yml -p=$(CURDIR)/data/pipes.yml
