#!/usr/bin/env bash

set -euo pipefail

if [[ $# -eq 0 ]]; then
  printf 'usage: %s <markdown-file>...\n' "$0" >&2
  exit 64
fi

timeout_seconds=${MERMAID_TIMEOUT_SECONDS:-120}
if ! [[ $timeout_seconds =~ ^[1-9][0-9]*$ ]]; then
  printf 'MERMAID_TIMEOUT_SECONDS must be a positive integer\n' >&2
  exit 64
fi

if [[ -n ${MMDC_BIN:-} ]]; then
  renderer=("$MMDC_BIN")
elif command -v mmdc >/dev/null 2>&1; then
  renderer=("$(command -v mmdc)")
else
  cli_version=${MERMAID_CLI_VERSION:-11.16.0}
  renderer=(npx --yes "@mermaid-js/mermaid-cli@${cli_version}")
fi

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/payment-book-mermaid.XXXXXX")
cleanup() {
  find "$work_dir" -type f -delete 2>/dev/null || true
  rmdir "$work_dir" 2>/dev/null || true
}
trap cleanup EXIT

render_one() {
  local input=$1
  local output=$2
  local log=$3
  local renderer_pid elapsed exit_code

  "${renderer[@]}" -i "$input" -o "$output" >"$log" 2>&1 &
  renderer_pid=$!
  elapsed=0

  while (( elapsed < timeout_seconds )); do
    if [[ -s $output ]] && grep -q '</svg>' "$output"; then
      # Some Chromium versions finish the SVG but hang during browser cleanup.
      # The complete closing tag is the artifact boundary; terminate only the
      # renderer that this function started and keep the successful evidence.
      if kill -0 "$renderer_pid" 2>/dev/null; then
        kill -TERM "$renderer_pid" 2>/dev/null || true
      fi
      wait "$renderer_pid" 2>/dev/null || true
      return 0
    fi

    # Mermaid can report a parse error while Chromium remains alive during
    # cleanup. Surface that error immediately instead of waiting for timeout.
    if [[ -s $log ]] && grep -q '^Error:' "$log"; then
      kill -TERM "$renderer_pid" 2>/dev/null || true
      wait "$renderer_pid" 2>/dev/null || true
      sed -n '1,160p' "$log" >&2
      return 1
    fi

    if ! kill -0 "$renderer_pid" 2>/dev/null; then
      if wait "$renderer_pid"; then
        exit_code=0
      else
        exit_code=$?
      fi
      if [[ $exit_code -eq 0 && -s $output ]] && grep -q '</svg>' "$output"; then
        return 0
      fi
      sed -n '1,160p' "$log" >&2
      if [[ $exit_code -eq 0 ]]; then
        printf 'Renderer exited without a complete SVG: %s\n' "$input" >&2
        return 1
      fi
      return "$exit_code"
    fi

    sleep 1
    elapsed=$((elapsed + 1))
  done

  kill -TERM "$renderer_pid" 2>/dev/null || true
  wait "$renderer_pid" 2>/dev/null || true
  printf 'Mermaid render timed out after %ss: %s\n' "$timeout_seconds" "$input" >&2
  sed -n '1,160p' "$log" >&2
  return 124
}

source_index=0
total=0
for source in "$@"; do
  if [[ ! -f $source ]]; then
    printf 'Markdown file does not exist: %s\n' "$source" >&2
    exit 66
  fi

  source_index=$((source_index + 1))
  block_count=$(awk -v out="$work_dir" -v prefix="$source_index" '
    /^```mermaid[[:space:]]*$/ {
      count++
      target = sprintf("%s/%s-%03d.mmd", out, prefix, count)
      inside = 1
      next
    }
    /^```[[:space:]]*$/ && inside {
      close(target)
      inside = 0
      next
    }
    inside { print > target }
    END {
      if (inside) exit 65
      print count + 0
    }
  ' "$source")

  if [[ $block_count -eq 0 ]]; then
    printf 'No Mermaid blocks: %s\n' "$source" >&2
    exit 65
  fi

  block=1
  while (( block <= block_count )); do
    input=$(printf '%s/%s-%03d.mmd' "$work_dir" "$source_index" "$block")
    output="${input%.mmd}.svg"
    log="${input%.mmd}.log"
    render_one "$input" "$output" "$log"
    block=$((block + 1))
    total=$((total + 1))
  done
  printf 'Mermaid OK: %s (%d blocks)\n' "$source" "$block_count"
done

printf 'Mermaid rendered: %d blocks\n' "$total"
