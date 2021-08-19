#! /usr/bin/env bash

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

log() {
  echo >&2 "$@"
}

cd "$SCRIPT_DIR"

packages_to_test=$(go list ./... | grep -v '/end-to-end-test$')
go test $packages_to_test

./end-to-end-test/run-tests.sh
