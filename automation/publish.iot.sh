#!/bin/sh

set -e

SELF=$0
REAL=$(realpath "$SELF")
BASE=$(dirname "$REAL")

VERSION=$(git describe --tags --dirty)

cd $BASE/..
docker login -u ${REGISTRY_USER} -p ${REGISTRY_TOKEN}
docker tag phytofy-amd64:latest ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}${VERSION}
docker tag phytofy-amd64:latest ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}latest
docker tag phytofy-arm32v7:latest ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}${VERSION}
docker tag phytofy-arm32v7:latest ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}latest
docker push ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}${VERSION}
docker push ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}latest
docker push ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}${VERSION}
docker push ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}latest
docker rmi phytofy-amd64:latest
docker rmi ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}${VERSION}
docker rmi ${REGISTRY_PREFIX}amd64${VERSION_SEPARATOR}latest
docker rmi phytofy-arm32v7:latest
docker rmi ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}${VERSION}
docker rmi ${REGISTRY_PREFIX}arm32v7${VERSION_SEPARATOR}latest
docker logout
