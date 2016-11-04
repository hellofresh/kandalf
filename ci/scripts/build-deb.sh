#!/usr/bin/env bash

set -e -u -x

# Install utility to build a DEB package
gem install fpm

# Define variables
APP_NAME="kandalf"
DIR_CI="./app-master/ci"
DIR_DEBIAN_SCRIPTS=${DIR_CI}/debian
DIR_DEBIAN_TMP="../deb"
DIR_OUT="./artifacts"
DIR_RESOURCES=${DIR_CI}/resources
VERSION=`${DIR_OUT}/kandalf -v | cut -d ' ' -f 3`
PKG_NAME="${APP_NAME}_${VERSION}_amd64.deb"

# Copy all needful files to a temp directory
mkdir ${DIR_DEBIAN_TMP}
mkdir -p ${DIR_DEBIAN_TMP}/etc/${APP_NAME}
mkdir -p ${DIR_DEBIAN_TMP}/usr/local/bin
install -m 644 ${DIR_RESOURCES}/config.yml ${DIR_DEBIAN_TMP}/etc/${APP_NAME}/config.yml
install -m 644 ${DIR_RESOURCES}/pipes.yml ${DIR_DEBIAN_TMP}/etc/${APP_NAME}/pipes.yml
install -m 755 ${DIR_OUT}/kandalf ${DIR_DEBIAN_TMP}/usr/local/bin

# Build DEB package
fpm -n ${APP_NAME} \
    -v ${VERSION} \
    -t deb \
    -s dir \
    -C ${DIR_DEBIAN_TMP} \
    -p ${DIR_OUT} \
    --config-files /etc/${APP_NAME} \
    --after-install ${DIR_DEBIAN_SCRIPTS}/postinst \
    --after-remove ${DIR_DEBIAN_SCRIPTS}/postrm \
    --deb-init ${DIR_DEBIAN_SCRIPTS}/${APP_NAME} \
    .

rm -rf ${DIR_DEBIAN_TMP}

# Upload package to a github
echo "Package uploaded"
