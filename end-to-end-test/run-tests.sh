#! /usr/bin/env bash

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

cd "$SCRIPT_DIR"

log() {
  echo >&2 "$@"
}

fail() {
  log "$@"
  exit 1
}

start-server-to-be-crawled() {
  # Clean up on exit
  trap 'kill $(jobs -p) >/dev/null 2>&1; rm -rf $tmp_dir' 0
  trap 'exit 2' SIGHUP SIGINT SIGQUIT SIGPIPE SIGTERM

  tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t "$SCRIPT_NAME")
  server_log_file="$tmp_dir/server.log"
  go run ./server.go > "$server_log_file" &

  max_tries=10
  i=0

  until grep -q 'Serving on' "$server_log_file"; do
    i=$(($i + 1))
    [[ $i -lt $max_tries ]] || fail 'Timed out waiting for server to start'
    sleep 0.5
  done

  server_url=$(sed 's/Serving on //' "$server_log_file")
}

start-server-to-be-crawled

expected_output="$server_url/
  -> $server_url/architecture
  -> $server_url/development

$server_url/architecture
  -> $server_url/
  -> $server_url/development

$server_url/development
  -> $server_url/
  -> $server_url/architecture"

(
  cd ..
  go build github.com/hilverd/sitemapper
)

actual_output=$(../sitemapper "$server_url/")

if diff -q >/dev/null <(echo "$expected_output") <(echo "$actual_output"); then
  log 'ok  	end-to-end tests'
else
  fail "FAIL: end-to-end tests. Expected output:

$expected_output

Actual output:

$actual_output"
fi
