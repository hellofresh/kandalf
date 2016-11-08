#!/usr/bin/env bash

set -e -u

. "`dirname "$0"`/_bootstrap.sh"

make bootstrap
make check
