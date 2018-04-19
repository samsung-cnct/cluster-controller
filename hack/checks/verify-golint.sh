#!/bin/bash

# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..

cd "${ROOT}"

golint=$(which golint)
if [[ ! -x "${golint}" ]]; then
  echo "could not find golint, please verify your GOPATH"
  exit 1
fi

source "${ROOT}/hack/common.sh"

for gofile in $(valid_go_files); do
  golint -set_exit_status "$gofile"
done
