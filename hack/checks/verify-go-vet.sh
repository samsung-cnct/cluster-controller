#!/bin/bash

# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..

cd "${ROOT}"

source "${ROOT}/hack/common.sh"

# gofmt exits with non-zero exit code if it finds a problem unrelated to
# formatting (e.g., a file does not parse correctly). Without "|| true" this
# would have led to no useful error message from gofmt, because the script would
# have failed before getting to the "echo" in the block below.
diff=$( echo `packages` | xargs go vet 2>&1) || true
if [[ -n "${diff}" ]]; then
  echo "${diff}"
  exit 1
fi