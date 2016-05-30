export GOPATH=$(CURDIR)/.go

APP_NAME = kandalf
OUTDIR=$(CURDIR)/out
VERSION=`$(OUTDIR)/$(APP_NAME) -v | cut -d ' ' -f 3`

$(OUTDIR)/$(APP_NAME): $(CURDIR)/src/main.go
	go build -o $(OUTDIR)/$(APP_NAME) $(CURDIR)/src/main.go
	chmod 0755 $(OUTDIR)/$(APP_NAME)

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

fmt:
	gofmt -s=true -w $(CURDIR)/src

run:
	go run -v $(CURDIR)/src/main.go -c=$(CURDIR)/data/config.yml -p=$(CURDIR)/data/pipes.yml
