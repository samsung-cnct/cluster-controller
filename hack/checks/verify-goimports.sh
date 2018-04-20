#!/bin/bash


# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..

cd "${ROOT}"

goimports=$(which goimports)
if [[ ! -x "${goimports}" ]]; then
  echo "could not find goimports, please verify your GOPATH"
  exit 1
fi

source "${ROOT}/hack/common.sh"

errors=$( echo `packages` | xargs ${goimports} -d 2>&1) || true
if [[ -n "${errors}" ]]; then
  echo "${errors}"
  exit 1
fi