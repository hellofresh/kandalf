#!/usr/bin/env bash

# Download latest Go
GO_VERSION=${GO_VERSION:="1.6.2"}
CLEANED_PATH=$(echo $PATH | sed -r 's|/(usr/local\|tmp)/go(/([0-9]\.)+[0-9])?/bin:||g')
CACHED_DOWNLOAD="${HOME}/cache/go${GO_VERSION}.linux-amd64.tar.gz"
PKG_NAME="kandalf_${CI_BRANCH}_amd64.deb"

export GOROOT="/tmp/go/${GO_VERSION}"
export PATH="${GOROOT}/bin:${CLEANED_PATH}"

mkdir -p "${GOROOT}"
wget --continue --output-document "${CACHED_DOWNLOAD}" "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz"
tar -xaf "${CACHED_DOWNLOAD}" --strip-components=1 --directory "${GOROOT}"
go version | grep ${GO_VERSION}

# Install FPM
gem install fpm

# Download the tool "github-release"
go get github.com/aktau/github-release

# Change the app version
sed -i -e "s/%app.version%/${CI_BRANCH}/g" src/main.go

# Build the binary
make dep-install
make deb

# Upload package to release assets
github-release upload \
    --user hellofresh \
    --repo kandalf \
    --tag ${CI_BRANCH} \
    --name ${PKG_NAME} \
    --file out/${PKG_NAME}
