FROM bitnami/minideb:buster
LABEL os=windows
LABEL arch=amd64

ENV GOVERSION="1.18.5"
ENV GOOS=windows
ENV GOARCH=amd64
ENV GOPATH=/go
ENV CGO_ENABLED=1
ENV CC=x86_64-w64-mingw32-gcc
ENV CXX=x86_64-w64-mingw32-g++
ENV PATH="/go/bin:/usr/local/go/bin:${PATH}"

# install build dependencies
RUN install_packages ca-certificates build-essential gcc-mingw-w64-x86-64 wget

# fix SSL issues down the line
RUN update-ca-certificates

# install golang
RUN set -ex; \
    osname=$(uname -s | tr '[:upper:]' '[:lower:]'); \
    hardware=$(dpkg --print-architecture); \
    goarchive=go$GOVERSION.$osname-$hardware.tar.gz; \
    wget -P /tmp -q https://go.dev/dl/$goarchive \
    && tar -C /usr/local -xzf /tmp/$goarchive
