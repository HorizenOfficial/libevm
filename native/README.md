# libevm

libevm implements a shared library to access a standalone instance of the go-ethereum EVM and its state storage layer StateDB and underlying LevelDB.

For simplicity all exported library functions take one parameter and return one value, which are all typed as C-strings and contain JSON.

## Build

To build both Linux and Windows binaries check the prerequisites below and execute the build script:
```sh
./build.sh
```

To build Linux and Windows binaries manually see the following sections.

### Prerequisites

- `go` version 1.21 or newer
- `gcc` GNU compiler
- `gcc-mingw-w64-x86-64` for cross-compilation to Windows
- `solc` Solidity compiler version 0.8.18 or newer
  - only required to run tests

### Linux

The project can be build with the standard Go tooling:
```sh
go build -buildmode c-shared -o bin/linux-x86-64/libevm.so
```

As defined in the file `go.mod`, the required minimum version of Go is `1.21`.

### Cross compile for Windows

To cross-compile for Windows the mingw compiler is required. Either have `gcc-mingw-w64-x86-64` installed locally or use the docker image to build, see `build.sh` for more information.

Example using the docker image:
```sh
docker build -t horizen/libevm:win32-x86-64 .
docker run --rm -v "$PWD":/libevm -w /libevm horizen/libevm:win32-x86-64 /bin/sh -c "go build -buildmode c-shared -o bin/win32-x86-64/libevm.dll"
```

## Tests

Some tests require smart contract code which is compiled during `go generate` using the Solidity compiler `solc`. Make sure to have `solc` installed and run `go generate ./...` before tests.

Example to run all tests:
```sh
go generate ./...
go test ./...
```
