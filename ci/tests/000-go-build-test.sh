#!/bin/bash
set -eo pipefail

retval=0

# Building jar package
echo "" && echo "=== Running go build to compile native libraries ===" && echo ""

cd native
./build.sh || retval="$?"

if [ ! -f './bin/linux-x86-64/libevm.so' ] || [ ! -f './bin/win32-x86-64/libevm.dll' ]; then
  echo "" && echo "=== Error: libevm libraries failed to compile. The build is going to fail. ===" && echo ""
  exit 1
fi

# Running go tests
echo "" && echo "=== Running go tests ===" && echo ""
go generate ./... || retval="$?"
go test ./... || retval="$?"

exit "$retval"

