#!/bin/bash

# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

# This script is really designed to use after gofmt was found to fail and you want to
# go ahead with suggestions. Its worth while having this hear in case we want to do
# something slightly different in the future.

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..
source "${ROOT}/hack/common.sh"

cd "${ROOT}"

goimports=$(which goimports)
if [[ ! -x "${goimports}" ]]; then
  echo "could not find goimports, please verify your GOPATH"
  exit 1
fi

for file in $(valid_go_files); do
    inf "${goimports} -w $file"
    ${goimports} -w $file
done
