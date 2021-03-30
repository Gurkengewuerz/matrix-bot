#!/usr/bin/env bash

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
CGO_CFLAGS="-I${SCRIPTPATH}/include/" CGO_LDFLAGS="-L${SCRIPTPATH}/lib/" go build