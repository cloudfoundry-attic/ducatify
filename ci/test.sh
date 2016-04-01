#!/bin/bash
set -e -u -x

mkdir -p go/src/github.com/cloudfoundry-incubator/
cp -a ducatify-source go/src/github.com/cloudfoundry-incubator/ducatify
cd go

export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin

cd src/github.com/cloudfoundry-incubator/ducatify

go install ./vendor/github.com/onsi/ginkgo/ginkgo

ginkgo -r -skipPackage vendor
