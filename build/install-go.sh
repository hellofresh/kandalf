#!/usr/bin/env bash

# Download latest Go
GO_VERSION=${GO_VERSION:="1.6.2"}
CLEANED_PATH=$(echo $PATH | sed -r 's|/(usr/local\|tmp)/go(/([0-9]\.)+[0-9])?/bin:||g')
CACHED_DOWNLOAD="${HOME}/cache/go${GO_VERSION}.linux-amd64.tar.gz"

export GOROOT="/tmp/go/${GO_VERSION}"
export PATH="${GOROOT}/bin:${CLEANED_PATH}"

mkdir -p "${GOROOT}"
wget --continue --output-document "${CACHED_DOWNLOAD}" "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz"
tar -xaf "${CACHED_DOWNLOAD}" --strip-components=1 --directory "${GOROOT}"
go version | grep ${GO_VERSION}
