#!/usr/bin/env bash

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
LD_LIBRARY_PATH="$SCRIPTPATH/lib" ./matrix-github-bot "$@"