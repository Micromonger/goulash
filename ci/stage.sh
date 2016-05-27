#!/bin/bash

set -xe

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

go get -u github.com/Masterminds/glide

source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  go get -u github.com/onsi/ginkgo/ginkgo
  ginkgo -p -randomizeAllSpecs action config handler slackapi
  mkdir -p stage
  GOOS=linux GOARCH=amd64 go build -o stage/goulash github.com/pivotalservices/goulash/cmd/goulash
  cp -R Procfile manifest.yml manifests stage
  git describe --abbrev=0 --tags > stage/tag
popd
