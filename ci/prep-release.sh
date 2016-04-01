#!/bin/bash
set -e -u -x

export OUT_BINARIES=$PWD/binaries
export OUT_NOTES=$PWD/release-notes

version=$(cat version/number)
echo "v${version}" > $OUT_NOTES/name

mkdir -p go/src/github.com/cloudfoundry-incubator/
cp -a ducatify-source go/src/github.com/cloudfoundry-incubator/ducatify
cd go

export GOPATH=$PWD
export PATH=$PATH:$GOPATH/bin

cd src/github.com/cloudfoundry-incubator/ducatify

for GOOS in linux darwin windows; do
  go build -o $OUT_BINARIES/ducatify-$GOOS &
done

wait

git config --global user.email "$GIT_USER_EMAIL"
git config --global user.name "$GIT_USER_NAME"

git rev-parse HEAD > $OUT_NOTES/commitish
