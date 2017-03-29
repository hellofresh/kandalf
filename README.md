<p align="center">
  <a href="https://hellofresh.com">
    <img width="120" src="https://www.hellofresh.de/images/hellofresh/press/HelloFresh_Logo.png">
  </a>
</p>

# Kandalf

> RabbitMQ to Kafka bridge

The main idea is to read messages from provided exchanges in [RabbitMQ](https://www.rabbitmq.com/) and send them to [Kafka](http://kafka.apache.org/).

Application uses intermediate permanent storage for keeping read messages in case of Kafka unavailability.

Service is written in Go language and can be build with go compiler of version 1.6 and above.

## Configuring

### Application configuration

Application is configured with environment variables or YAML config file.

#### Environment variables

* `LOG_LEVEL` - Logging verbosity level, see [logrus](https://github.com/Sirupsen/logrus#level-logging) for details (_default_: `info`)
* `RABBIT_DSN` - RabbiMQ server DSN
* `STORAGE_DSN` - Permanent storage DSN, where Scheme is storage type. The following storage types are currently supported:
  * [Redis](https://redis.io/) - requires, `key` as DSN query parameter as redis storage key, e.g. `redis://localhost:6379/?key=kandalf`
* `KAFKA_BROKERS` - Kafka brokers comma-separated list, e.g. `192.168.0.1:9092,192.168.0.2:9092`
* `KAFKA_MAX_RETRY` - Total number of times to retry sending a message to Kafka (_default_: `5`)
* `KAFKA_PIPES_CONFIG` - Path to RabbitMQ-Kafka bridge mappings config, see details below (_default_: `/etc/kandalf/conf/pipes.yml`)
* `STATS_DSN` - Stats host, see [hellofresh/stats-go](https://github.com/hellofresh/stats-go#usage) for usage details.
* `STATS_PREFIX` - Stats prefix, see [hellofresh/stats-go](https://github.com/hellofresh/stats-go#usage) for usage details.
* `WORKER_CYCLE_TIMEOUT` - Main application bridge worker cycle timeout to avoid CPU overload, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `2s`)
* `WORKER_CACHE_SIZE` - Max messages number that we store in memory before trying to publish to Kafka (_default_: `10`)
* `WORKER_CACHE_FLUSH_TIMEOUT` - Max amount of time we store messages in memory before trying to publish to Kafka, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `5s`)
* `WORKER_STORAGE_READ_TIMEOUT` - Timeout between attempts of reading persisted messages from storage, to publish them to Kafka, must be at least 2x greater than `WORKER_CYCLE_TIMEOUT`, must be valid [duration string](https://golang.org/pkg/time/#ParseDuration) (_default_: `10s`)

#### Config file

You can use `-c <file_path>` parameter to load application settings from YAML file. Config should have the following structure:

```yaml
log_level: "info" # same as env LOG_LEVEL
rabbit_dsn: "amqp://user:password@rmq" # same as env RABBIT_DSN
storage_dsn: "redis://redis.local/?key=storage:key" # same as env STORAGE_DSN
kafka:
  brokers: # same as env KAFKA_BROKERS
    - "192.0.0.1:9092"
    - "192.0.0.2:9092"
  max_retry: 5 # same as env KAFKA_MAX_RETRY
  pipes_config: "/etc/kandalf/conf/pipes.yml" # same as env KAFKA_PIPES_CONFIG
stats:
  dsn: "statsd.local:8125" # same as env STATS_DSN
  prefix: "kandalf" # same as env STATS_PREFIX
worker:
  cycle_timeout: "2s" # same as env WORKER_CYCLE_TIMEOUT
  cache_size: 10 # same as env WORKER_CACHE_SIZE
  cache_flush_timeout: "5s" # same as env WORKER_CACHE_FLUSH_TIMEOUT
  storage_read_timeout: "10s" # same as env WORKER_STORAGE_READ_TIMEOUT
```

### Pipes configuration

The rules, defining which messages should be send to which Kafka topics, are defined in "[pipes.yml](./ci/assets/pipes.yml)" and is called "pipes".

Each pipe should contain following keys:

* **kafka_topic** _str_ — name of the topic in Kafka where message will be sent;
* **rabbitmq_exchange_name** _str_ — name of the exchange in RabbitMQ;
* **rabbitmq_routing_key** _str_ — routing key for exchange;
* **rabbitmq_queue_name** _str_ — the name of RabbitMQ queue to read messages from.

## How to build a binary on a local machine

1. Make sure that you have `go` and `make` utility installed on your machine;
2. Run: `make` to install all required dependencies and build binaries;
3. Binaries for Linux and MacOS X would be in `./dist/`.

## How to run service in a docker environment

For testing and development you can use [`docker-compose`](./docker-compose.yml) file with all the required services.

For production you can use minimalistic prebuilt [hellofresh/kandalf](quay.io/hellofresh/kandalf) image as base image or mount pipes configuration volume to `/etc/kandalf/conf/`.

## Todo

* [x] Handle dependencies in a proper way (gvt, glide or smth.)
* [ ] Tests

## Contributing

To start contributing, please check [CONTRIBUTING](CONTRIBUTING.md).

## Documentation

* Kandalf Docs: https://hellofresh.gitbooks.io/kandalf
* Kandalf Go Docs: https://godoc.org/github.com/hellofresh/kandalf
* Go lang: https://golang.org/
