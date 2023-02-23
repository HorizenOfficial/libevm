#!/bin/sh

set -ex

# preparation
go generate ./...

echo "# building for linux"
# build library
go build -buildmode c-shared -o "bin/linux-x86-64/libevm.so"
# build all test executables
for i in $(go list ./...); do go test -c -o "bin/linux-x86-64/tests/$i.test" $i; done

echo "# building docker image for cross-compilation to windows"
docker build -t horizen/libevm:win32-x86-64 .

echo "# building for windows"
docker run --rm -v "$PWD":/libevm -w /libevm horizen/libevm:win32-x86-64 /bin/sh -c \
  "set -ex && go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll && for i in \$(go list ./...); do go test -c -o \"bin/win32-x86-64/tests/\$i.exe\" \$i; done"

# dump symbols of shared object
nm -g bin/linux-x86-64/libevm.so
winedump -j export bin/win32-x86-64/libevm.dll
