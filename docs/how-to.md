---
id: how-to
title: How to guidelines
description: Main How to guidelines 
---

## How to build a binary on a local machine

1. Make sure you have `go` and `make` utility installed on your machine;
2. Run: `make` to install all required dependencies and build binaries;
3. Binaries for Linux and MacOS X would be in `./dist/`.

## How to run service in a docker environment

For testing and development you can use [`docker-compose`](./docker-compose.yml) file with all the required services.

For production you can use minimalistic prebuilt [hellofreshtech/kandalf](https://hub.docker.com/r/hellofreshtech/kandalf/tags) image as base image or mount pipes configuration volume to `/etc/kandalf/conf/`.
