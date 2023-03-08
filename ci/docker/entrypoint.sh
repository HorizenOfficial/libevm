#!/bin/bash
set -euo pipefail

# Add local zenbuilder user, either use LOCAL_USER_ID:LOCAL_GRP_ID
# if set via environment or fallback to 9001:9001
USER_ID="${LOCAL_USER_ID:-9001}"
GRP_ID="${LOCAL_GRP_ID:-9001}"
if [ "$USER_ID" != "0" ]; then
    export USERNAME=zenbuilder
    getent group "$GRP_ID" &> /dev/null || groupadd -g "$GRP_ID" "$USERNAME"
    id -u "$USERNAME" &> /dev/null || useradd --shell /bin/bash -u "$USER_ID" -g "$GRP_ID" -o -c "" -m "$USERNAME"
    CURRENT_UID="$(id -u "$USERNAME")"
    CURRENT_GID="$(id -g "$USERNAME")"
    export HOME=/home/"$USERNAME"
    if [ "$USER_ID" != "$CURRENT_UID" ] || [ "$GRP_ID" != "$CURRENT_GID" ]; then
        echo "WARNING: User with differing UID ${CURRENT_UID}/GID ${CURRENT_GID} already exists, most likely this container was started before with a different UID/GID. Re-create it to change UID/GID."
    fi
else
    export USERNAME=root
    export HOME=/root
    CURRENT_UID="$USER_ID"
    CURRENT_GID="$GRP_ID"
    echo "WARNING: Starting container processes as root. This has some security implications and goes against docker best practice."
fi

# Installing extra dependencies only for TESTS stage
if [ -n "${TESTS:-}" ]; then
  # Installing extra dependencies
  echo "" && echo "=== Installing extra dependencies for the build ===" && echo ""

  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -yq software-properties-common
  # Repo for solc
  add-apt-repository -y ppa:ethereum/ethereum
  # Repo for golang-go version >= 1.18
  add-apt-repository -y ppa:longsleep/golang-backports
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -yq gcc gcc-mingw-w64-x86-64 libc6-dev solc golang-"${GOLANG_VERSION}"
  export PATH="/usr/lib/go-${GOLANG_VERSION}/bin:$PATH"
fi

# Print information
echo "" && echo "=== Environment Info ===" && echo ""
java --version
echo
if command -v go; then
  go version
  echo
fi
lscpu
echo
free -h
echo
echo "Username: $USERNAME, HOME: $HOME, UID: $CURRENT_UID, GID: $CURRENT_GID"

if [ "$USERNAME" = "root" ]; then
  exec /build/ci/docker/entrypoint_setup_gpg.sh "$@"
else
  exec gosu "$USERNAME" /build/ci/docker/entrypoint_setup_gpg.sh "$@"
fi

