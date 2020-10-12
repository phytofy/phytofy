#!/bin/sh

SELF=$0
REAL=$(realpath "$SELF")
BASE=$(dirname "$REAL")

cd "$BASE/.."
docker build -t phytofy-cli:latest -f docker/CLI.Dockerfile --build-arg GH_API_USER --build-arg GH_API_TOKEN .
