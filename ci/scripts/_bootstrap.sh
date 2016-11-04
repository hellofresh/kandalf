#!/usr/bin/env bash

mkdir -p ${GOPATH}/${PROJECT_SRC}
cp -r ${RESOURCE}/* ${GOPATH}/${PROJECT_SRC}
cd ${GOPATH}/${PROJECT_SRC}
