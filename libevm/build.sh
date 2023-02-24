#!/bin/sh

set -ex

# delete output folder
sudo rm -rf bin

# preparation
go generate ./...

echo "# building for linux"
# build library
go build -buildmode c-shared -o "bin/linux-x86-64/libevm.so"
# build all test executables
#for i in $(go list ./...); do go test -c -o "bin/linux-x86-64/tests/$i.test" $i; done

#echo "# building docker image for cross-compilation to windows"
#docker build -t horizen/libevm:win32-x86-64 .

echo "# building for windows"
docker run --rm -v "$PWD":/libevm -w /libevm horizen/libevm:win32-x86-64 /bin/sh -c \
  "go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll"

# dump symbols of shared object
#nm -g bin/linux-x86-64/libevm.so
winedump -j export bin/win32-x86-64/libevm.dll

# change ownership for all directories created by the docker container
sudo chown -R gigo bin

# move binaries to the EVM package
cp -r bin/linux-x86-64 ../evm/target/classes/
cp -r bin/win32-x86-64 ../evm/target/classes/
