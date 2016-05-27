#!/bin/bash

set -xe

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

go get -u github.com/Masterminds/glide

stage_dir="$PWD/stage"
source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  GOOS=linux GOARCH=amd64 go build -o $stage_dir/goulash github.com/pivotalservices/goulash/cmd/goulash
  cp -R Procfile manifest.yml manifests $stage_dir
  git describe --abbrev=0 --tags > $stage_dir/tag
popd
