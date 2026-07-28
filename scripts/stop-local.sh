#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="${ROOT_DIR}/.runs/local"

# Collect pid + all descendants (depth-first leaves first).
collect_tree() {
  local pid="$1"
  local child
  while read -r child; do
    [[ -n "${child}" ]] || continue
    collect_tree "${child}"
  done < <(pgrep -P "${pid}" 2>/dev/null || true)
  echo "${pid}"
}

signal_pids() {
  local sig="$1"
  shift
  local pid
  for pid in "$@"; do
    kill "-${sig}" "${pid}" 2>/dev/null || true
  done
}

any_alive() {
  local pid
  for pid in "$@"; do
    if kill -0 "${pid}" 2>/dev/null; then
      return 0
    fi
  done
  return 1
}

stop_service() {
  local name="$1"
  local pid_file="$2"

  if [[ ! -f "${pid_file}" ]]; then
    echo "${name} is not running."
    return
  fi

  local pid
  pid="$(cat "${pid_file}")"
  if ! kill -0 "${pid}" 2>/dev/null; then
    rm -f "${pid_file}"
    echo "${name} is not running."
    return
  fi

  local pgid
  pgid="$(ps -o pgid= -p "${pid}" 2>/dev/null | tr -d '[:space:]' || true)"

  local -a tree=()
  while read -r child; do
    [[ -n "${child}" ]] || continue
    tree+=("${child}")
  done < <(collect_tree "${pid}")

  # Prefer process-group kill when this pid leads its own session/group.
  if [[ -n "${pgid}" && "${pgid}" == "${pid}" ]]; then
    kill -TERM -- "-${pgid}" 2>/dev/null || true
  else
    signal_pids TERM "${tree[@]}"
  fi

  local _
  for _ in {1..20}; do
    if ! any_alive "${tree[@]}"; then
      rm -f "${pid_file}"
      echo "${name} stopped."
      return
    fi
    sleep 0.25
  done

  if [[ -n "${pgid}" && "${pgid}" == "${pid}" ]]; then
    kill -KILL -- "-${pgid}" 2>/dev/null || true
  fi
  signal_pids KILL "${tree[@]}"

  rm -f "${pid_file}"
  echo "${name} was force-stopped."
}

stop_service "UI" "${RUN_DIR}/ui.pid"
stop_service "API" "${RUN_DIR}/api.pid"
