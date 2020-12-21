#!/bin/sh

set -e

SELF=$0
REAL=$(realpath "$SELF")
BASE=$(dirname "$REAL")

cd "$BASE/.."
docker run --rm -v ${PWD}:/target phytofy-cli:latest cp '/app/phytofy-cli*' /target/
