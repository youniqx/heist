#!/usr/bin/env bash
set -o nounset
set -o errexit
set -o pipefail
set -o xtrace

setup_fake_gopath() {
  FAKE_GOPATH="$(mktemp -d)"
  FAKE_REPOPATH="${FAKE_GOPATH}/src/${PKG_BASE}"
  mkdir -p "$(dirname "${FAKE_REPOPATH}")"
  ln -s "${REPO_ROOT}" "${FAKE_REPOPATH}"
  export GOPATH="${FAKE_GOPATH}"
  cd "${FAKE_REPOPATH}"
}

main() {
  REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"
  PKG_BASE=github.com/youniqx/heist

  cd "${REPO_ROOT}"

  # turn off module mode before running the generators
  # https://github.com/kubernetes/code-generator/issues/69

  export GO111MODULE="off"

  setup_fake_gopath

  for version in "${@}"; do

    "${REPO_ROOT}/bin/client-gen" -v 9 \
        --input-base "${PKG_BASE}/pkg/apis" \
        --clientset-name heist \
        -i "./pkg/apis/heist.youniqx.com/${version}/" \
        --input "heist.youniqx.com/${version}" \
        --output-package "${PKG_BASE}/pkg/client/heist.youniqx.com/${version}/clientset" \
        --go-header-file hack/boilerplate.go.txt \
        -o "${GOPATH}/src"

    "${REPO_ROOT}/bin/lister-gen" -v 5 \
        -i "${PKG_BASE}/pkg/apis/heist.youniqx.com/${version}" \
        --output-package "${PKG_BASE}/pkg/client/heist.youniqx.com/${version}/listers" \
        --go-header-file hack/boilerplate.go.txt \
        -o "${GOPATH}/src"

  done

  export GO111MODULE="on"
  cd "${REPO_ROOT}"
}

main "${@}"
