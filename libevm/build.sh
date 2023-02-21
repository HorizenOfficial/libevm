#!/bin/sh

set -ex

badArguments() {
  echo "please provide a build target, i.e. either linux or windows"
  exit 1
}

if [ $# != 1 ]; then
  badArguments
fi

# preparation
go generate ./...

case "$1" in
"windows")
  echo "building for windows"
  out=bin/windows-amd64
  exe_ext=.exe
  lib_ext=.dll
  docker run --rm -v "$PWD/bin/.cache":/go -v "$PWD":/libevm -w /libevm horizen/libevm:windows-amd64 /bin/sh -c \
    "set -ex && go build -buildmode c-shared -o bin/windows-amd64/libevm.dll && for i in \$(go list ./...); do go test -c -o \"bin/windows-amd64/tests/\$i.exe\" \$i; done"
  ;;
"linux")
  echo "building for linux"
  out=bin/linux-x86-64
  exe_ext=""
  lib_ext=.so
  # build library
  go build -buildmode c-shared -o "${out}/libevm${lib_ext}"
  # build all test executables
  for i in $(go list ./...); do go test -c -o "${out}/tests/${i}${exe_ext}" $i; done
  ;;
*)
  badArguments
  ;;
esac
