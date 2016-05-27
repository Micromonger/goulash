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
popd
