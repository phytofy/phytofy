#!/bin/sh

set -e

SELF=$0
REAL=$(realpath "$SELF")
BASE=$(dirname "$REAL")

cd "$BASE/.."
docker run --rm -v ${PWD}:/target phytofy-cli:latest /bin/sh -c 'cp /app/release/* /target/'
