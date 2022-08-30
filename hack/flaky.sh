#!/usr/bin/env bash

set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__file="${__dir}/$(basename "${BASH_SOURCE[0]}")"
__base="$(basename "${__file}" .sh)"
__root="$(cd "$(dirname "${__dir}")" && pwd)"

TEST_TARGET=""
NUMBER_OF_RUNS=""

printUsage() {
  cat <<EOF
Usage: $0 [options] test-target number-of-runs
  This script helps find/rerun flaky test targets. It can be configured with
  multiple optional options, and two required command arguments.

  Examples
    runs the makefile target test_controllers_vaulttransitkey 3 times
      $0 test_controllers_vaulttransitkey 3

    runs the makefile target test_controllers_vaulttransitkey 3 times but fails
    on the first encountered error
      $0 --fail-fast test_controllers_vaulttransitkey 3


  Required Arguments
    test-target       The Makefile target that will be run.
    number-of-runs    The number of times the target will be run.

  Options
    -h | --help       Show this help
    -f | --fail-fast  Fails as soon as an error is encountered during a test run.
EOF
}

parseOptionsAndParams() {
  while true; do
    case "${1}" in
      -h|--help)
        printUsage
        exit 0
        ;;
      -f|--fail-fast)
        set -o errexit
        shift
        ;;
      *)
        TEST_TARGET="${1}"
        NUMBER_OF_RUNS="${2}"
        shift 2
        break
    esac
  done
}

checkParams() {
  if [[ "${TEST_TARGET}" == "" ]] ||  [[ "${NUMBER_OF_RUNS}" == "" ]]; then
    printUsage

    echo
    echo
    echo "required arguments missing (TEST_TARGET: ${TEST_TARGET}, NUMBER_OF_RUNS: ${NUMBER_OF_RUNS})"
    exit 1
  fi
}

runTests() {
  local testTarget="${TEST_TARGET}"
  local runs="${NUMBER_OF_RUNS}"
  local failed=0

  for i in $(seq "${runs}"); do
    echo
    echo
    echo
    echo "Run ${i}/${runs}"
    make "${testTarget}" || ((failed++))
    make clean
  done

  echo "Failed: ${failed}/${runs}"

  if [[ "${failed}" != "0" ]]; then
    exit 1
  fi
}

main() {
  parseOptionsAndParams "$@"
  checkParams "$@"
  runTests "$@"
}

main "$@"
