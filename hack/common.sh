#!/bin/bash

# Some useful colors.
declare -r color_start="\033["
declare -r color_red="${color_start}0;31m"
declare -r color_yellow="${color_start}0;33m"
declare -r color_green="${color_start}0;32m"
declare -r color_norm="${color_start}0m"

SILENT=true

function warn {
  echo -e "${color_yellow}WARNING: $1${color_norm}"
}

function error {
  echo -e "${color_red}ERROR: $1${color_norm}"
}

function inf {
  echo -e "${color_green}$1${color_norm}"
}

packages() {
    // TODO
  echo ""
}

valid_go_files() {
  git ls-files "**/*.go" "*.go" | grep -v -e "vendor" -e "pkg/client"
}


function print-failed-tests {
  echo -e "========================"
  error "FAILED TESTS"
  echo -e "========================"
  for t in ${FAILED_TESTS[@]}; do
      error "${t}"
  done
}

function print-failed-go-test-package {
  echo -e "========================"
  error "FAILED GO TESTS"
  echo -e "========================"
  for t in ${1}; do
      error "go test ${t}"
  done
}


function run-cmd {
  if ${SILENT}; then
    "$@" &> /dev/null
  else
    "$@"
  fi
}