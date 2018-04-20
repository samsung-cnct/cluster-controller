#!/bin/bash

# from http://github.com/kubernetes/kubernetes/hack/verify-gofmt.sh

# This script is really designed to use after gofmt was found to fail and you want to
# go ahead with suggestions. Its worth while having this hear in case we want to do
# something slightly different in the future.

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/../..

cd "${ROOT}"

gofmt=$(which gofmt)
if [[ ! -x "${gofmt}" ]]; then
  echo "could not find gofmt, please verify your GOPATH"
  exit 1
fi

source "${ROOT}/hack/common.sh"

echo $(valid_go_files)
for file in $(valid_go_files); do
    inf "${gofmt} -s -w $file"
    ${gofmt} -s -w $file
done
