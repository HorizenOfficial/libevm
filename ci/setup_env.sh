#!/bin/bash

set -eo pipefail

export CONTAINER_PUBLISH="false"
PUBLISH_BUILD="${PUBLISH_BUILD:-false}"
prod_release="false"
mapfile -t prod_release_br_list < <(echo "${PROD_RELEASE_BRANCHES}" | tr " " "\n")

pom_version="$(xpath -q -e '/project/version/text()' ./evm/pom.xml)"

if [ -z "${TRAVIS_TAG}" ]; then
  echo "TRAVIS_TAG:                           No TAG"
else
  echo "TRAVIS_TAG:                           $TRAVIS_TAG"
fi
echo "Production release branch(es):        ${prod_release_br_list[*]}"
echo "./evm/pom.xml version:                $pom_version"

# Functions
# Functions
function import_gpg_keys() {
  # shellcheck disable=SC2145
  printf "%s\n" "Tagged build, fetching keys:" "${@}" ""
  # shellcheck disable=SC2207
  declare -r my_arr=( $(echo "${@}" | tr " " "\n") )

  for key in "${my_arr[@]}"; do
    echo "Importing key: ${key}"
    gpg -v --batch --keyserver hkps://keys.openpgp.org --recv-keys "${key}" ||
    gpg -v --batch --keyserver hkp://keyserver.ubuntu.com --recv-keys "${key}" ||
    gpg -v --batch --keyserver hkp://pgp.mit.edu:80 --recv-keys "${key}" ||
    gpg -v --batch --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys "${key}" ||
    echo -e "${key} can not be found on GPG key servers. Please upload it to at least one of the following GPG key servers:\nhttps://keys.openpgp.org/\nhttps://keyserver.ubuntu.com/\nhttps://pgp.mit.edu/"
  done
}

function check_signed_tag() {
  # Checking if git tag signed by the maintainers
  if git verify-tag -v "${1}"; then
    echo "${1} is a valid signed tag"
    return 0
  fi
  echo "Git tag's = ${1} gpg signature is NOT valid. The build is not going to be released..."
  return 1
}

function release_prep() {
  echo "" && echo "=== ${1} release build ===" && echo ""
  echo "Fetching maven gpg signing keys."
  curl -sLH "Authorization: token ${GITHUB_TOKEN}" -H "Accept: application/vnd.github.v3.raw" "${MAVEN_KEY_ARCHIVE_URL}" |
    openssl enc -d -aes-256-cbc -md sha256 -pass pass:"${MAVEN_KEY_ARCHIVE_PASSWORD}" |
    tar -xzf- -C "${HOME}"
  export CONTAINER_PUBLISH="true"
}

# empty key.asc file in case we're not signing
touch "${HOME}/key.asc"

if [ -n "${TRAVIS_TAG}" ] && [ "${PUBLISH_BUILD}" = "true" ]; then
  # checking if MAINTAINER_KEYS is set
  if [ -z "${MAINTAINER_KEYS}" ]; then
    echo "Error: MAINTAINER_KEYS variable is not set. Make sure to set it up for release build.  Exiting ...."
    exit 1
  fi

  # shellcheck disable=SC2155
  export GNUPGHOME="$(mktemp -d 2>/dev/null || mktemp -d -t 'GNUPGHOME')"
  import_gpg_keys "${MAINTAINER_KEYS}"

  # Checking git tag gpg signature requirement
  if ( check_signed_tag "${TRAVIS_TAG}" ); then
    # Prod vs dev release
    for release_branch in "${prod_release_br_list[@]}"; do
      if ( git branch -r --contains "${TRAVIS_TAG}" | grep -xqE ". origin\/${release_branch}$" ); then
        # Checking format of production release pom version
        if ! [[ "${pom_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$ ]]; then
          echo "Error: Aborting, pom version is in the wrong format.  Exiting ..."
          exit 1
        fi

        # Checking Github tag format
        if ! [[ "${TRAVIS_TAG}" == "${pom_version}" ]]; then
          echo -e "\nError: Tag format differs from the pom file version."
          echo -e "Github tag name: ${TRAVIS_TAG}\nPom file version: ${pom_version}.\nPublish stage is NOT going to run. Exiting ..."
          exit 1
        fi

        # Announcing PROD release
        release_prep Production
        prod_release="true"
      fi
    done
    if [ "${prod_release}" = "false" ]; then
      # Checking if package version matches DEV release version
      if ! [[ "${pom_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?(-SNAPSHOT){1}$ ]]; then
        echo "Error: Aborting, pom version is in the wrong format.  Exiting ..."
        exit 1
      fi

      # Checking Github tag format
      if ! [[ "${TRAVIS_TAG}" =~ "${pom_version}"[0-9]*$ ]]; then
        echo "Error: Aborting, tag format differs from the pom file.  Exiting ..."
        exit 1
      fi

      # Announcing DEV release
      release_prep Development
    fi
  fi
fi

# unset credentials if not publishing
if [ "${CONTAINER_PUBLISH}" = "false" ]; then
  export CONTAINER_OSSRH_JIRA_USERNAME=""
  export CONTAINER_OSSRH_JIRA_PASSWORD=""
  export CONTAINER_GPG_KEY_NAME=""
  export CONTAINER_GPG_PASSPHRASE=""
  unset CONTAINER_OSSRH_JIRA_USERNAME
  unset CONTAINER_OSSRH_JIRA_PASSWORD
  unset CONTAINER_GPG_KEY_NAME
  unset CONTAINER_GPG_PASSPHRASE
  echo "" && echo "=== NOT a release build ===" && echo ""
fi

# unset credentials after use
export GITHUB_TOKEN=""
export MAVEN_KEY_ARCHIVE_URL=""
export MAVEN_KEY_ARCHIVE_PASSWORD=""
unset GITHUB_TOKEN
unset MAVEN_KEY_ARCHIVE_URL
unset MAVEN_KEY_ARCHIVE_PASSWORD

set +eo pipefail
