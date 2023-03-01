#!/bin/sh

set -e

resourcesDir=../evm/src/main/resources
linuxPrefix=linux-x86-64
windowsPrefix=win32-x86-64

# delete output directory
rm -rf bin

echo "# building for linux"
go build -buildmode c-shared -o "bin/$linuxPrefix/libevm.so"

echo "# building for windows"
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -buildmode c-shared -o bin/$windowsPrefix/libevm.dll

# alternative: build using docker container
#docker build -t horizen/libevm:$windowsPrefix .
#docker run --rm -v "$PWD":/libevm -w /libevm horizen/libevm:$windowsPrefix /bin/sh -c \
#  "go build -buildmode c-shared -o bin/$windowsPrefix/libevm.dll"

# dump symbols of shared object
#nm -g bin/$linuxPrefix/libevm.so
#winedump -j export bin/$windowsPrefix/libevm.dll

echo "# copy binaries to the EVM package"
mkdir -p $resourcesDir/$linuxPrefix
cp bin/$linuxPrefix/libevm.so $resourcesDir/$linuxPrefix/libevm.so
mkdir -p $resourcesDir/$windowsPrefix
cp bin/$windowsPrefix/libevm.dll $resourcesDir/$windowsPrefix/libevm.dll
