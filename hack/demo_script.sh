#!/usr/bin/env bash
# Script run inside asciinema to produce the demo recording.
# Not for production use — uses the mock aquadirector in hack/.

export PATH="$(cd "$(dirname "$0")" && pwd):$PATH"

# Simulate a prompt
PS1="$ "

_type() {
  local text="$1"
  for ((i=0; i<${#text}; i++)); do
    printf '%s' "${text:$i:1}"
    sleep 0.04
  done
}

_run() {
  echo -n "$ "
  _type "$1"
  echo
  sleep 0.3
  eval "$1"
}

sleep 0.5
_run "aquadirector dashboard"
sleep 2
_run "aquadirector sensor status --output json | jq '{ph, temperature_c, sg}'"
sleep 1.5
_run "aquadirector alerts check"
sleep 1.5
