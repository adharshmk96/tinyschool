#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PLAYWRIGHT_DIR="${SCRIPT_DIR}/../playwright"

if [[ ! -d "${PLAYWRIGHT_DIR}/node_modules" ]]; then
  echo "Installing Playwright dependencies..."
  (cd "${PLAYWRIGHT_DIR}" && bun install)
fi

echo "Taking Tiny School screenshots..."
(cd "${PLAYWRIGHT_DIR}" && bun run screenshots)
echo "Light screenshots saved to ${PLAYWRIGHT_DIR}/output/light"
echo "Dark screenshots saved to ${PLAYWRIGHT_DIR}/output/dark"
