#!/bin/sh

set -ex

# delete output folder
rm -rf bin

# preparation
go generate ./...

echo "# building for linux"
go build -buildmode c-shared -o "bin/linux-x86-64/libevm.so"

echo "# building for windows"
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll

# dump symbols of shared object
#nm -g bin/linux-x86-64/libevm.so
winedump -j export bin/win32-x86-64/libevm.dll

# move binaries to the EVM package
cp -r bin/linux-x86-64 ../evm/target/classes/
cp -r bin/win32-x86-64 ../evm/target/classes/
