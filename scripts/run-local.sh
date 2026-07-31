#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="${ROOT_DIR}/.runs/local"
API_PID_FILE="${RUN_DIR}/api.pid"
UI_PID_FILE="${RUN_DIR}/ui.pid"
API_HOST="${TINYSCHOOL_API_HOST:-127.0.0.1}"
API_PORT="${TINYSCHOOL_API_PORT:-8080}"
UI_HOST="${TINYSCHOOL_UI_HOST:-127.0.0.1}"
UI_PORT="${TINYSCHOOL_UI_PORT:-3000}"
PUBLIC_HOST="${TINYSCHOOL_PUBLIC_HOST:-localhost}"

is_running() {
  local pid_file="$1"
  [[ -f "${pid_file}" ]] && kill -0 "$(cat "${pid_file}")" 2>/dev/null
}

# Start command in a new session so it survives the launcher shell and
# can be stopped as a whole process group (parent + children).
start_detached() {
  local log_file="$1"
  local pid_file="$2"
  shift 2

  python3 - "$log_file" "$pid_file" "$@" <<'PY'
import os
import sys

log_file, pid_file, *cmd = sys.argv[1:]
if not cmd:
    raise SystemExit("missing command")

pid = os.fork()
if pid == 0:
    os.setsid()
    with open(log_file, "wb") as log:
        os.dup2(log.fileno(), 1)
        os.dup2(log.fileno(), 2)
    os.closerange(3, 1024)
    os.execvp(cmd[0], cmd)
    raise SystemExit(127)

with open(pid_file, "w", encoding="utf-8") as handle:
    handle.write(str(pid))
PY
}

start_service() {
  local name="$1"
  local work_dir="$2"
  local log_file="$3"
  local pid_file="$4"
  shift 4

  if is_running "${pid_file}"; then
    echo "${name} is already running (PID $(cat "${pid_file}"))."
    return
  fi

  rm -f "${pid_file}"
  (
    cd "${work_dir}"
    start_detached "${log_file}" "${pid_file}" "$@"
  )

  sleep 1
  if ! is_running "${pid_file}"; then
    echo "${name} failed to start. See ${log_file}." >&2
    exit 1
  fi

  echo "${name} started (PID $(cat "${pid_file}"))."
}

mkdir -p "${RUN_DIR}" "${RUN_DIR}/go-cache"

start_service \
  "API" \
  "${ROOT_DIR}/tinyschool-api" \
  "${RUN_DIR}/api.log" \
  "${API_PID_FILE}" \
  env \
    GOCACHE="${RUN_DIR}/go-cache" \
    TINYSCHOOL_API_ADDRESS="${API_HOST}:${API_PORT}" \
    TINYSCHOOL_DB_PATH="${RUN_DIR}/tinyschool.db" \
    TINYSCHOOL_APP_BASE_URL="http://${PUBLIC_HOST}:${UI_PORT}" \
    go run .

start_service \
  "UI" \
  "${ROOT_DIR}/tinyschool-ui" \
  "${RUN_DIR}/ui.log" \
  "${UI_PID_FILE}" \
  env \
    NUXT_PUBLIC_API_BASE="http://${PUBLIC_HOST}:${API_PORT}/api/v1" \
    NUXT_PUBLIC_APP_VERSION="${TINYSCHOOL_APP_VERSION:-$(git -C "${ROOT_DIR}" describe --tags --always 2>/dev/null || echo dev)}" \
    bun run dev --host "${UI_HOST}" --port "${UI_PORT}"

echo "Tiny School is starting at http://${PUBLIC_HOST}:${UI_PORT}"
echo "Logs: ${RUN_DIR}"
