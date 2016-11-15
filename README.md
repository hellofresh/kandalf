<p align="center">
  <a href="https://hellofresh.com">
    <img width="120" src="https://www.hellofresh.de/images/hellofresh/press/HelloFresh_Logo.png">
  </a>
</p>

# kandalf also known as "RabbitMQ to Kafka bridge"

The main idea is to read messages from provided exchanges in [RabbitMQ](https://www.rabbitmq.com/) and send them to [Kafka](http://kafka.apache.org/).

## Implementation

This service is written in Go language and can be build with go compiler of version 1.6 and above.

The rules, defining which messages should be send to which Kafka topics, are defined in "[pipes.yml](./ci/resources/pipes.yml)" and is called "pipes".

Each pipe should contain following keys.
* **kafka_topic** _str_ — name of the topic in kafka where message will be sent;
* **rabbitmq_exchange_name** _str_ — name of the exchange in RabbitMQ;
* **rabbitmq_routing_key** _str_ — routing key for exchange;
* **rabbitmq_queue_name** _str_ — the name of the kandalf's queue. Used to make multiple kandalf instances consume messages from the same queue.

## How to build a binary on a local machine

1. Make sure that you have `make` utility installed on your machine;
2. Run: `make bootstrap` to install [glide](https://glide.sh) and dependencies;
3. Run: `make` to build binary;
4. Binaries for Linux and MacOS X would be in `./out/`.

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

## How to contribute

Just follow instructions placed in [CONTRIBUTING.md](./CONTRIBUTING.md).

## Todo

* [x] Handle dependencies in a proper way (gvt, glide or smth.)
* [ ] Tests

--------
HelloFresh - More Than Food.
