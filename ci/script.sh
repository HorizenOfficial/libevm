#!/bin/bash
# shellcheck disable=SC2046

set -euo pipefail

command -v docker &> /dev/null && have_docker="true" || have_docker="false"
# absolute path to project from relative location of this script
workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
# defaults if not provided via env
GOLANG_VERSION="${GOLANG_VERSION:-1.18}"
DOCKER_ORG="${DOCKER_ORG:-zencash}"
IMAGE_NAME="${IMAGE_NAME:-sc-ci-base}"
IMAGE_TAG="${IMAGE_TAG:-bionic_jdk-11_latest}"
image="${DOCKER_ORG}/${IMAGE_NAME}:${IMAGE_TAG}"
export DISABLE_JAVADOC_LINT="${DISABLE_JAVADOC_LINT:-false}"
export PUBLISH_BUILD="${PUBLISH_BUILD:-false}"

if [ "${1}" = "tests" ]; then
  # run tests in docker or natively
  if [ -n "${TESTS:-}" ]; then
    if [ "$have_docker" = "true" ]; then
      if [ -n "${DOCKER_READONLY_USERNAME:-}" ] && [ -n "${DOCKER_READONLY_PASSWORD:-}" ]; then
        echo "$DOCKER_READONLY_PASSWORD" | docker login -u "$DOCKER_READONLY_USERNAME" --password-stdin
      fi
      docker run --rm -t -v "$workdir":/build -w /build \
        --entrypoint /build/ci/docker/entrypoint.sh \
        -v "${HOME}"/key.asc:/key.asc \
        -e PUBLISH_BUILD \
        -e TESTS \
        -e GOLANG_VERSION \
        -e LOCAL_USER_ID="$(id -u)" \
        -e LOCAL_GRP_ID="$(id -g)"\
        "${image}" ci/run_tests.sh
    else
      cd "${workdir:-.}" && ci/run_tests.sh
    fi
  fi
elif [ "${1}" = "release" ]; then
  if [ "$have_docker" = "true" ]; then
    if [ -n "${DOCKER_READONLY_USERNAME:-}" ] && [ -n "${DOCKER_READONLY_PASSWORD:-}" ]; then
      echo "$DOCKER_READONLY_PASSWORD" | docker login -u "$DOCKER_READONLY_USERNAME" --password-stdin
    fi
    docker run --rm -t -v "$workdir":/build -w /build \
      --entrypoint /build/ci/docker/entrypoint.sh \
      -v "${HOME}"/key.asc:/key.asc \
      -e DISABLE_JAVADOC_LINT \
      -e PUBLISH_BUILD \
      -e TRAVIS_TAG \
      -e LOCAL_USER_ID="$(id -u)" \
      -e LOCAL_GRP_ID="$(id -g)"\
      $(env | grep -E '^CONTAINER_' | sed -n '/^[^\t]/s/=.*//p' | sed '/^$/d' | sed 's/^/-e /g' | tr '\n' ' ') \
      "${image}" ci/publish.sh
  else
    cd "${workdir:-.}" && ci/publish.sh
  fi
fi