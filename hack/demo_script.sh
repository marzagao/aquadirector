#!/usr/bin/env bash
# Script run inside asciinema to produce the demo recording.
# Not for production use — uses the mock aquadirector in hack/.

export PATH="$(cd "$(dirname "$0")" && pwd):$PATH"

PS1="$ "

# Type text character by character with varying delay for a natural feel.
_type() {
  local text="$1"
  local i ch delay
  for ((i=0; i<${#text}; i++)); do
    ch="${text:$i:1}"
    printf '%s' "$ch"
    # Jitter: most characters 40-90ms, occasional longer pauses on spaces
    if [[ "$ch" == " " ]]; then
      delay="0.0$(( RANDOM % 6 + 6 ))"   # 60-110ms after spaces
    else
      delay="0.0$(( RANDOM % 5 + 4 ))"   # 40-80ms for letters
    fi
    sleep "$delay"
  done
}

_run() {
  echo -n "$ "
  _type "$1"
  echo
  sleep 0.4
  eval "$1"
}

sleep 0.5
_run "aquadirector dashboard"
sleep 3
_run "aquadirector sensor status --output json | jq '{ph, temperature_c, sg}'"
sleep 3
_run "aquadirector alerts check"
sleep 2
