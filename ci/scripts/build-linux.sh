#!/usr/bin/env bash

set -e -u -x

pushd .

. "`dirname "$0"`/_bootstrap.sh"

VERSION="0.0.1"

# Change the app version
sed -i -e "s/%app.version%/${VERSION}/g" cmd/kandalf/main.go

make bootstrap
make build-linux

popd

cp ${GOPATH}/${PROJECT_SRC}/out/linux/kandalf ./build/
