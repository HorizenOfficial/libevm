#!/bin/sh

set -ex

# preparation
go generate ./...

echo "# building for linux"
# build library
go build -buildmode c-shared -o "bin/linux-x86-64/libevm.so"
# build all test executables
for i in $(go list ./...); do go test -c -o "bin/linux-x86-64/tests/$i.test" $i; done

echo "# building for windows"
docker run --rm -v "$PWD/bin/.cache":/go -v "$PWD":/libevm -w /libevm horizen/libevm:windows-amd64 /bin/sh -c \
  "set -ex && go build -buildmode c-shared -o bin/windows-amd64/libevm.dll && for i in \$(go list ./...); do go test -c -o \"bin/windows-amd64/tests/\$i.exe\" \$i; done"
