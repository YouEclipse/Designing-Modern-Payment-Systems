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

setup_timeout_seconds=${MERMAID_SETUP_TIMEOUT_SECONDS:-300}
if ! [[ $setup_timeout_seconds =~ ^[1-9][0-9]*$ ]]; then
  printf 'MERMAID_SETUP_TIMEOUT_SECONDS must be a positive integer\n' >&2
  exit 64
fi

npm_fetch_timeout_ms=${MERMAID_NPM_FETCH_TIMEOUT_MS:-60000}
if ! [[ $npm_fetch_timeout_ms =~ ^[1-9][0-9]*$ ]]; then
  printf 'MERMAID_NPM_FETCH_TIMEOUT_MS must be a positive integer\n' >&2
  exit 64
fi

npm_fetch_retries=${MERMAID_NPM_FETCH_RETRIES:-1}
if ! [[ $npm_fetch_retries =~ ^[0-9]+$ ]]; then
  printf 'MERMAID_NPM_FETCH_RETRIES must be a non-negative integer\n' >&2
  exit 64
fi

book_root=$(cd "$(dirname "$0")/.." && pwd -P)
temporary_root="$book_root/.validation-tmp"
mkdir -p "$temporary_root"

work_dir=$(mktemp -d "$temporary_root/payment-book-mermaid.XXXXXX")
cleanup() {
  find "$work_dir" -type f -delete 2>/dev/null || true
  rmdir "$work_dir" 2>/dev/null || true
}
trap cleanup EXIT

export NPM_CONFIG_CACHE="$temporary_root/npm-cache"
export NPM_CONFIG_FETCH_TIMEOUT="$npm_fetch_timeout_ms"
export NPM_CONFIG_FETCH_RETRIES="$npm_fetch_retries"
export NPM_CONFIG_UPDATE_NOTIFIER=false
export NPM_CONFIG_FUND=false
export NPM_CONFIG_AUDIT=false
export PUPPETEER_CACHE_DIR="$temporary_root/puppeteer-cache"
mkdir -p "$NPM_CONFIG_CACHE" "$PUPPETEER_CACHE_DIR"

if [[ -z ${PUPPETEER_EXECUTABLE_PATH:-} ]]; then
  for browser in \
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
    "/Applications/Chromium.app/Contents/MacOS/Chromium" \
    /usr/bin/google-chrome /usr/bin/chromium /usr/bin/chromium-browser; do
    if [[ -x $browser ]]; then
      export PUPPETEER_EXECUTABLE_PATH="$browser"
      export PUPPETEER_SKIP_DOWNLOAD=true
      break
    fi
  done
fi

stop_process() {
  local pid=$1

  kill -TERM "$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
}

force_stop_process() {
  local pid=$1

  kill -TERM "$pid" 2>/dev/null || true
  if kill -0 "$pid" 2>/dev/null; then
    kill -KILL "$pid" 2>/dev/null || true
  fi
  wait "$pid" 2>/dev/null || true
}

verify_renderer_version() {
  local output="$work_dir/renderer-version"
  local log="$work_dir/renderer-version.log"
  local renderer_pid elapsed exit_code actual_version

  TMPDIR="$temporary_root" "${version_command[@]}" --version >"$output" 2>"$log" &
  renderer_pid=$!
  elapsed=0

  while (( elapsed < setup_timeout_seconds )); do
    if ! kill -0 "$renderer_pid" 2>/dev/null; then
      if wait "$renderer_pid"; then
        exit_code=0
      else
        exit_code=$?
      fi
      if [[ $exit_code -ne 0 ]]; then
        sed -n '1,160p' "$log" >&2
        return "$exit_code"
      fi
      actual_version=$(tr -d '\r\n' <"$output")
      if [[ $actual_version != "$cli_version" ]]; then
        printf 'Mermaid CLI version mismatch: expected %s, got %s\n' "$cli_version" "${actual_version:-<empty>}" >&2
        sed -n '1,160p' "$log" >&2
        return 65
      fi
      printf '%s\n' "$actual_version" >"$work_dir/mermaid-cli-version"
      mv "$work_dir/mermaid-cli-version" "$temporary_root/mermaid-cli-version"
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done

  force_stop_process "$renderer_pid"
  printf 'Mermaid CLI setup timed out after %ss\n' "$setup_timeout_seconds" >&2
  sed -n '1,160p' "$log" >&2
  return 124
}

pinned_cli_version=11.16.0
cli_version=$pinned_cli_version
if [[ -n ${MMDC_BIN:-} ]]; then
  cli_version=${MERMAID_CLI_VERSION:-$pinned_cli_version}
  if [[ $MMDC_BIN == */* ]]; then
    resolved_renderer=$MMDC_BIN
  else
    resolved_renderer=$(command -v -- "$MMDC_BIN" || true)
  fi
  if [[ -z ${resolved_renderer:-} || ! -x $resolved_renderer ]]; then
    printf 'MMDC_BIN is not executable: %s\n' "$MMDC_BIN" >&2
    exit 69
  fi
  renderer=("$resolved_renderer")
  version_command=("$resolved_renderer")
else
  if [[ -n ${MERMAID_CLI_VERSION:-} && $MERMAID_CLI_VERSION != "$pinned_cli_version" ]]; then
    printf 'MERMAID_CLI_VERSION can override the pinned version only with explicit MMDC_BIN\n' >&2
    exit 64
  fi
  renderer=(npx --yes "@mermaid-js/mermaid-cli@${pinned_cli_version}")
  version_command=(npx --yes "@mermaid-js/mermaid-cli@${pinned_cli_version}")
fi

# Resolve/install and verify the exact renderer version before the per-diagram
# timeout starts. Setup is separately bounded and all npm/Puppeteer/temp files
# stay below the active book repository.
verify_renderer_version

render_one() {
  local input=$1
  local output=$2
  local log=$3
  local renderer_pid elapsed exit_code

  TMPDIR="$temporary_root" "${renderer[@]}" -i "$input" -o "$output" >"$log" 2>&1 &
  renderer_pid=$!
  elapsed=0

  while (( elapsed < timeout_seconds )); do
    if [[ -s $output ]] && grep -q '</svg>' "$output"; then
      # Some Chromium versions finish the SVG but hang during browser cleanup.
      # The complete closing tag is the artifact boundary; terminate only the
      # renderer that this function started and keep the successful evidence.
      if kill -0 "$renderer_pid" 2>/dev/null; then
        stop_process "$renderer_pid"
      else
        wait "$renderer_pid" 2>/dev/null || true
      fi
      return 0
    fi

    # Mermaid can report a parse error while Chromium remains alive during
    # cleanup. Surface that error immediately instead of waiting for timeout.
    if [[ -s $log ]] && grep -q '^Error:' "$log"; then
      stop_process "$renderer_pid"
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

  force_stop_process "$renderer_pid"
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
