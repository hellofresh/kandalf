#!/usr/bin/env bash

# Remove the "v" prefix from deb-package name
VERSION=$(echo ${CI_BRANCH} | tr -d 'v')
PKG_NAME="kandalf_${VERSION}_amd64.deb"

# Install FPM
gem install fpm

# Download the tool "github-release"
go get github.com/aktau/github-release

# Change the app version
sed -i -e "s/%app.version%/${VERSION}/g" src/main.go

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
