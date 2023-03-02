#!/bin/sh

set -ex

# delete output directory
rm -rf bin

echo "# building for linux"
go build -buildmode c-shared -o "bin/linux-x86-64/libevm.so"

echo "# building for windows"
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll

# alternative: build using docker container
#docker build -t horizen/libevm:win32-x86-64 .
#docker run --rm -v "$PWD":/libevm -w /libevm horizen/libevm:win32-x86-64 /bin/sh -c \
#  "go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll"

# dump symbols of shared object
#nm -g bin/linux-x86-64/libevm.so
#winedump -j export bin/win32-x86-64/libevm.dll

echo "# copy binaries to the EVM package"
cp -r bin/linux-x86-64 ../evm/target/classes/
cp -r bin/win32-x86-64 ../evm/target/classes/
