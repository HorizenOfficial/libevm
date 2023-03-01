# libevm

This projects implements a shared library and a Java wrapper for it. This library provides access to a standalone version of the go-ethereum EVM and its state storage layer StateDB and underlying LevelDB.

## Build

### shared library

For details on how to build the shared library see [libevm/README.md](libevm/README.md).

### java wrapper

The java wrapper is build via Maven: execute `mvn package` in the `evm` directory.
