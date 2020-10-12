#!/bin/sh

SELF=$0
REAL=$(realpath "$SELF")
BASE=$(dirname "$REAL")

cd "$BASE/.."
docker build -t phytofy-amd64:latest -f docker/IoT.amd64.Dockerfile --build-arg GH_API_USER --build-arg GH_API_TOKEN .
docker build -t phytofy-arm32v7:latest -f docker/IoT.arm32v7.Dockerfile --build-arg GH_API_USER --build-arg GH_API_TOKEN .
