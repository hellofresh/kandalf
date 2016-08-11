#!/usr/bin/env bash
# This script is used only with Codeship to build deb-package
# and attatch it to release on Github

# Remove the "v" prefix from deb-package name
VERSION=$(echo ${CI_BRANCH} | tr -d 'v')
PKG_NAME="kandalf_${VERSION}_amd64.deb"

# Install FPM
gem install fpm

# Download the tool "github-release"
go get github.com/aktau/github-release

# Change the app version
sed -i -e "s/%app.version%/${VERSION}/g" main.go

# Build the binary
make bootstrap
make deb

# Upload package to release assets
github-release upload \
    --user hellofresh \
    --repo kandalf \
    --tag ${CI_BRANCH} \
    --name ${PKG_NAME} \
    --file build/out/${PKG_NAME}
