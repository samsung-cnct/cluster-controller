#!/bin/bash

# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..

cd "${ROOT}"

gocyclo=$(which gocyclo)
if [[ ! -x "${gocyclo}" ]]; then
  echo "could not find gocyclo, please verify your GOPATH"
  echo "https://github.com/fzipp/gocyclo"
  exit 1
fi

source "${ROOT}/hack/common.sh"


for gofile in $(valid_go_files); do
  gocyclo -over 15 "$gofile"
done
