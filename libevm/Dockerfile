FROM rwgrim/msys2-cross
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
ENV PKG_CONFIG_PATH=/windows/mingw64/lib/pkgconfig
ENV MSYS2_ARCH=x86_64

# install build dependencies
RUN set -ex \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        gcc-mingw-w64-x86-64 \
    && rm -rf /var/lib/apt/lists/*

# install golang
RUN set -ex; \
    osname=$(uname -s | tr '[:upper:]' '[:lower:]'); \
    hardware=$(dpkg --print-architecture); \
    goarchive=go$GOVERSION.$osname-$hardware.tar.gz; \
    wget -P /tmp -q https://go.dev/dl/$goarchive \
    && tar -C /usr/local -xzf /tmp/$goarchive
