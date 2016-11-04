#!/usr/bin/env bash

set -e -u -x

pushd .

. "`dirname "$0"`/_bootstrap.sh"

VERSION="0.0.1"

# Change the app version
sed -i -e "s/%app.version%/${VERSION}/g" cmd/kandalf/main.go

make bootstrap
make build-linux
go get github.com/aktau/github-release

popd

cp ${GOPATH}/${PROJECT_SRC}/out/linux/kandalf ./artifacts/
cp ${GOPATH}/bin/github-release ./artifacts/
