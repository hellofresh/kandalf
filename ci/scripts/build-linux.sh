#!/usr/bin/env bash

set -e -u -x

pushd .

VERSION=$(cat ./semantic-version/version)

. "`dirname "$0"`/_bootstrap.sh"

# Change the app version
sed -i -e "s/%app.version%/${VERSION}/g" cmd/kandalf/main.go

make bootstrap
make build-linux

popd

cp ${GOPATH}/${PROJECT_SRC}/out/linux/kandalf ./build/
