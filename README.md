# kandalf also known as "RabbitMQ to Kafka bridge"

The main idea is to read all messages from all queues in RabbitMQ and send them to Kafka.

## Implementation

This service is written in Go language and can be build with go compiler of version 1.6 and above.

To get all messages from RabbitMQ without any exception "[Firehose Tracer](https://www.rabbitmq.com/firehose.html)" should be enabled:

```bash
rabbitmqctl trace_on
```

The rules, defining which messages should be send to which Kafka topics, are defined in "[pipes.yml](./build/resources/pipes.yml)" and is called "pipes".

Each pipe should contain at least name of the topic in Kafka and at least one of the following rules related to metadata from RabbitMQ: "exchange_name", "routed_queue" and "routing_key".

* **exchange_name** _str_ — name of the exchange to which the message was published;
* **routed_queue** _str_ — all the routed queues of a message must match this pattern;
* **routing_key** _str_ — all routing keys of a message must match this pattern.

All rules supports patterns in Unix shell style described [here](https://golang.org/pkg/path/).

The rules matching uses logical conjunction. In other words to match the proper Kafka topic, all rules in pipe should be fully met.

Also the **priority** of pipe is important, pipes with higher priority are more relevant and important than the lower ones.

## How to build a binary on a local machine

1. Make sure that you have `make` utility installed on your machine;
2. Run: `make bootstrap` to install [glide](https://glide.sh) and dependencies;
3. Run: `make` to build binary;
4. Binaries for Linux and MacOS X would be in `build/out/`.

## Hot to build deb-package

1. Make sure that you have `make` utility installed on your machine;
2. Make sure that you have [Effing Package Management](https://github.com/jordansissel/fpm) installed on your machine;
3. Run: `make bootstrap` to install [glide](https://glide.sh) and dependencies;
4. Run: `make deb`

## How to run service in a docker environment

1. Make sure that you have `make` utility installed on your machine;
2. Make sure that you have `docker` and `docker-compose` installed.
3. Run: `make docker-up-env`
4. Run: `make`
5. Run: `make docker-run`

## Todo

* [x] Handle dependencies in a proper way (gvt, glide or smth.)
* [ ] More tests
