#!/bin/bash

set -xe

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

go get -u github.com/Masterminds/glide

stage_dir="$PWD/stage"
source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  mkdir -p $stage_dir/release
  GOOS=linux GOARCH=amd64 go build -o $stage_dir/release/goulash github.com/pivotalservices/goulash/cmd/goulash
  cp -f Procfile manifests/* manifests/.* $stage_dir/release
  git describe --abbrev=0 --tags > $stage_dir/tag
popd
