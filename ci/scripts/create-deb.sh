#!/usr/bin/env bash

set -e -u

CWD=$(pwd)
VERSION=$(cat ./version/version)
APP_NAME=kandalf

# Define variables
DIR_ARTIFACTS="${CWD}/artifacts"
DIR_DEBIAN_SCRIPTS="${CWD}/source-code/ci/assets/debian"
DIR_DEBIAN_TMP="${CWD}/deb"
PKG_NAME="${APP_NAME}_amd64.deb"

# Copy all needful files to a temp directory
mkdir -p ${DIR_DEBIAN_TMP}/etc/${APP_NAME}/conf
mkdir -p ${DIR_DEBIAN_TMP}/usr/local/bin
install -m 644 ${CWD}/source-code/ci/assets/config.yml ${DIR_DEBIAN_TMP}/etc/${APP_NAME}/conf/config.yml
install -m 644 ${CWD}/source-code/ci/assets/pipes.yml ${DIR_DEBIAN_TMP}/etc/${APP_NAME}/conf/pipes.yml
install -m 755 ${DIR_ARTIFACTS}/${APP_NAME} ${DIR_DEBIAN_TMP}/usr/local/bin

# Build DEB package
fpm --name ${APP_NAME} \
    --output-type deb \
    --version ${VERSION} \
    --input-type dir \
    --chdir ${DIR_DEBIAN_TMP} \
    --package ${DIR_ARTIFACTS} \
    --maintainer engineering@hellofresh.com \
    --config-files /etc/${APP_NAME}/conf \
    --after-install ${DIR_DEBIAN_SCRIPTS}/postinst \
    --after-remove ${DIR_DEBIAN_SCRIPTS}/postrm \
    --deb-init ${DIR_DEBIAN_SCRIPTS}/${APP_NAME} \
    .

rm -rf ${DIR_DEBIAN_TMP}
