<p align="center">
  <a href="https://hellofresh.com">
    <img width="120" src="https://www.hellofresh.de/images/hellofresh/press/HelloFresh_Logo.png">
  </a>
</p>

# Kandalf

[![Build Status](https://github.com/hellofresh/kandalf/workflows/Release/badge.svg)](https://github.com/hellofresh/kandalf/actions?query=workflow%3ARelease)
[![codecov](https://codecov.io/gh/hellofresh/kandalf/branch/master/graph/badge.svg)](https://codecov.io/gh/hellofresh/kandalf)
[![GoDoc](https://godoc.org/github.com/hellofresh/kandalf?status.svg)](https://godoc.org/github.com/hellofresh/kandalf)
[![Go Report Card](https://goreportcard.com/badge/github.com/hellofresh/kandalf)](https://goreportcard.com/report/github.com/hellofresh/kandalf)

## Note

As of version 0.7 docker images migrated to [Docker Hub](#how-to-run-service-in-a-docker-environment)

---

> RabbitMQ to Kafka bridge

The main idea is to read messages from provided exchanges in [RabbitMQ](https://www.rabbitmq.com/) and send them to [Kafka](http://kafka.apache.org/).

Application uses intermediate permanent storage for keeping read messages in case of Kafka unavailability.

Service is written in Go language and can be build with go compiler of version 1.14 and above.

## Configuring

### Application configuration

Application is configured with environment variables or config files of different formats - JSON, TOML, YAML, HCL, and Java properties.

By default it tries to read config file from `/etc/kandalf/conf/config.<ext>` and `./config.<ext>`. You can change the path using `-c <file_path>` or `--config <file_path>` application parameters. If file is not found config loader does fallback to reading config values from environment variables.

#### Environment variables

* `RABBIT_DSN` - RabbiMQ server DSN
* `STORAGE_DSN` - Permanent storage DSN, where Scheme is storage type. The following storage types are currently supported:
  * [Redis](https://redis.io/) - requires, `key` as DSN query parameter as redis storage key, e.g. `redis://localhost:6379/?key=kandalf`
* `LOG_*` - Logging settings, see [hellofresh/logging-go](https://github.com/hellofresh/logging-go#configuration) for details
* `KAFKA_BROKERS` - Kafka brokers comma-separated list, e.g. `192.168.0.1:9092,192.168.0.2:9092`
* `KAFKA_MAX_RETRY` - Total number of times to retry sending a message to Kafka (_default_: `5`)
* `KAFKA_PIPES_CONFIG` - Path to RabbitMQ-Kafka bridge mappings config, see details below (_default_: `/etc/kandalf/conf/pipes.yml`)
* `STATS_DSN` - Stats host, see [hellofresh/stats-go](https://github.com/hellofresh/stats-go#usage) for usage details.
* `STATS_PREFIX` - Stats prefix, see [hellofresh/stats-go](https://github.com/hellofresh/stats-go#usage) for usage details.
* `STATS_PORT` - Stats port, used only for `prometheus` metrics, metrics are exposed on `localhost:<port>/metrics` (_default_: `8080`).
* `WORKER_CYCLE_TIMEOUT` - Main application bridge worker cycle timeout to avoid CPU overload, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `2s`)
* `WORKER_CACHE_SIZE` - Max messages number that we store in memory before trying to publish to Kafka (_default_: `10`)
* `WORKER_CACHE_FLUSH_TIMEOUT` - Max amount of time we store messages in memory before trying to publish to Kafka, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `5s`)
* `WORKER_STORAGE_READ_TIMEOUT` - Timeout between attempts of reading persisted messages from storage, to publish them to Kafka, must be at least 2x greater than `WORKER_CYCLE_TIMEOUT`, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `10s`)
* `WORKER_STORAGE_MAX_ERRORS` - Max storage read errors in a row before worker stops trying reading in current read cycle. Next read cycle will be in `WORKER_STORAGE_READ_TIMEOUT` interval. (_default_: `10`)

#### Config file (YAML example)

Config should have the following structure:

```yaml
logLevel: "info"                                    # same as env LOG_LEVEL
rabbitDSN: "amqp://user:password@rmq"               # same as env RABBIT_DSN
storageDSN: "redis://redis.local/?key=storage:key"  # same as env STORAGE_DSN
kafka:
  brokers:                                          # same as env KAFKA_BROKERS
    - "192.0.0.1:9092"
    - "192.0.0.2:9092"
  maxRetry: 5                                       # same as env KAFKA_MAX_RETRY
  pipesConfig: "/etc/kandalf/conf/pipes.yml"        # same as env KAFKA_PIPES_CONFIG
stats:
  dsn: "statsd.local:8125"                          # same as env STATS_DSN
  prefix: "kandalf"                                 # same as env STATS_PREFIX
worker:
  cycleTimeout: "2s"                                # same as env WORKER_CYCLE_TIMEOUT
  cacheSize: 10                                     # same as env WORKER_CACHE_SIZE
  cacheFlushTimeout: "5s"                           # same as env WORKER_CACHE_FLUSH_TIMEOUT
  storageReadTimeout: "10s"                         # same as env WORKER_STORAGE_READ_TIMEOUT
  storageMaxErrors: 10                              # same as env WORKER_STORAGE_MAX_ERRORS
```

You can find sample config file in [assets/config.yml](./assets/config.yml).

### Pipes configuration

The rules, defining which messages should be send to which Kafka topics, are defined in Kafka Pipes Config file and are called "pipes". Each pipe has the following structure:

```yaml
- kafkaTopic: "loyalty"                                # name of the topic in Kafka where message will be sent
  rabbitExchangeName: "customers"                      # name of the exchange in RabbitMQ
  rabbitTransientExchange: false                       # determines if the exchange should be declared as durable or transient
  rabbitRoutingKey: "badge.received"                   # routing key for exchange
  rabbitQueueName: "kandalf-customers-badge.received"  # the name of RabbitMQ queue to read messages from
  rabbitDurableQueue: true                             # determines if the queue should be declared as durable
  rabbitAutoDeleteQueue: false                         # determines if the queue should be declared as auto-delete
```

You can find sample Kafka Pipes Config file in [assets/pipes.yml](./assets/pipes.yml).

## How to build a binary on a local machine

1. Make sure you have `go` and `make` utility installed on your machine;
2. Run: `make` to install all required dependencies and build binaries;
3. Binaries for Linux and MacOS X would be in `./dist/`.

## How to run service in a docker environment

For testing and development you can use [`docker-compose`](./docker-compose.yml) file with all the required services.

For production you can use minimalistic prebuilt [hellofresh/kandalf](https://hub.docker.com/r/hellofresh/kandalf/tags) image as base image or mount pipes configuration volume to `/etc/kandalf/conf/`.

## Todo

* [x] Handle dependencies in a proper way (gvt, glide or smth.)
* [ ] Tests

## Contributing

To start contributing, please check [CONTRIBUTING](CONTRIBUTING.md).
