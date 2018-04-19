#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE}")/..
source "${ROOT}/hack/common.sh"


# Collect Failed tests in this Array , initialize to nil
FAILED_TESTS=()

function run-check {
  local -r check=$1
  local -r runner=$2

  echo -e "Verifying ${check}"
  local start=$(date +%s)
  run-cmd "${runner}" "${check}" && tr=$? || tr=$?
  local elapsed=$(($(date +%s) - ${start}))
  if [[ ${tr} -eq 0 ]]; then
    echo -e "${color_green}SUCCESS${color_norm}  ${check}\t${elapsed}s"
  else
    echo -e "${color_red}FAILED${color_norm}   ${check}\t${elapsed}s"
    ret=1
    FAILED_TESTS+=(${check})
  fi
}

while getopts ":v" opt; do
  case ${opt} in
    v)
      SILENT=false
      ;;
    \?)
      echo "Invalid flag: -${OPTARG}" >&2
      exit 1
      ;;
  esac
done

if ${SILENT} ; then
  echo "Running in silent mode, run with -v if you want to see script logs."
fi


ret=0
run-check "${ROOT}/hack/checks/verify-go-vet.sh" bash
run-check "${ROOT}/hack/checks/verify-gofmt.sh" bash
run-check "${ROOT}/hack/checks/verify-goimports.sh" bash
run-check "${ROOT}/hack/checks/verify-golint.sh" bash
run-check "${ROOT}/hack/checks/verify-gocyclo.sh" bash


if [[ ${ret} -eq 1 ]]; then
    print-failed-tests
fi
exit ${ret}