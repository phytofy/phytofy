# Copyright (c) 2020 OSRAM; Licensed under the MIT license.

# Build API
FROM node:14.17.4-alpine3.14 as BuildAPI

COPY ./api /app/api
WORKDIR /app

RUN npm install -g openapi-types@9.1.0 swagger-cli@4.0.4 && \
    swagger-cli bundle -o /app/api/hw0.json /app/api/hw0.yaml && \
    swagger-cli bundle -o /app/api/hw1.json /app/api/hw1.yaml && \
    swagger-cli bundle -o /app/api/api.json /app/api/api.yaml


# Build JS licensing info
FROM node:14.17.4-alpine3.14 as BuildLicensingJS

COPY ./ui /app/ui

RUN cd /app/ui && \
    npm ci && \
    npm install -g license-checker@25.0.1 && \
    license-checker --json > /app/ui/licensing.js.txt


# Build Go licensing info
FROM golang:1.16.6-alpine3.14 AS BuildLicensingGo

COPY ./core /app/core

RUN cd /app/core && \
    go list -m all > /app/core/licensing.go.txt


# Build licensing info
FROM python:3.10.0-alpine3.14 as BuildLicensing

ARG GH_API_USER
ARG GH_API_TOKEN

ENV GH_API_USER=$GH_API_USER
ENV GH_API_TOKEN=$GH_API_TOKEN

COPY ./automation /app/automation
COPY --from=BuildLicensingJS /app/ui /app/ui
COPY --from=BuildLicensingGo /app/core /app/core

RUN python /app/automation/licenses_js.py /app/ui/licensing.js.txt > /tmp/licenses.js.fragment
RUN python /app/automation/licenses_go.py /app/core/licensing.go.txt > /tmp/licenses.go.fragment
RUN cat /tmp/licenses.js.fragment /tmp/licenses.go.fragment > /tmp/licenses.fragment
RUN python /app/automation/substitute.py /tmp/licenses.fragment /app/ui/src/views/PhytofyInformation.vue THIRD_PARTY_LICENSES
RUN rm -rf /app/ui/node_modules


# Build UI
FROM node:14.17.4-alpine3.14 as BuildUI

COPY --from=BuildLicensing /app/ui /app/ui
COPY --from=BuildAPI /app/api/hw0.json /app/ui/public/hw0.json
COPY --from=BuildAPI /app/api/hw1.json /app/ui/public/hw1.json
COPY --from=BuildAPI /app/api/api.json /app/ui/public/api.json
WORKDIR /app

RUN cd /app/ui && \
    npm ci && \
    npm run build


# Build App
FROM golang:1.16.6-alpine3.14 AS BuildApp

RUN apk update && apk add --no-cache git

COPY ./core /app/core
COPY --from=BuildUI /app/ui/dist /app/core/assets
WORKDIR /app/core

RUN go get -d -v && \
    go get -u golang.org/x/lint/golint && \
    golint -set_exit_status ./... && \
    rm /app/core/assets/.keep && \
    mkdir /app/release && \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /app/release/phytofy-cli.windows-amd64.exe && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/release/phytofy-cli.linux-amd64 && \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o /app/release/phytofy-cli.linux-arm32v7 && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o /app/release/phytofy-cli.macos-amd64
