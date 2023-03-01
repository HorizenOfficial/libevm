#!/bin/bash
set -eEuo pipefail

root_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/../.." &> /dev/null && pwd )"

CONTENT="$(xmllint --format --encode UTF-8 "${root_dir}/evm/pom.xml")"
echo "${CONTENT}" > evm/pom.xml

SETTINGS_CONTENT="$(xmllint --format --encode UTF-8 "${root_dir}/ci/mvn_settings.xml")"
echo "${SETTINGS_CONTENT}" > ci/mvn_settings.xml


######
# The END
######
echo "" && echo "=== Pom xml file(s) syntax was successfully validated ===" && echo ""

exit 0