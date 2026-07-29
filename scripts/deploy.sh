#!/usr/bin/env bash

set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_HOST="${DEPLOY_HOST:-147.93.97.228}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/tinyschool}"
DEPLOY_DOMAIN="${DEPLOY_DOMAIN:-tinyschool.${DEPLOY_HOST}.nip.io}"
DEPLOY_SSH_KEY="${DEPLOY_SSH_KEY:-}"
REGISTRY="${REGISTRY:-ghcr.io}"
REGISTRY_USER="${REGISTRY_USER:-}"
REGISTRY_TOKEN="${REGISTRY_TOKEN:-}"
IMAGE_UI="${IMAGE_UI:-}"
IMAGE_API="${IMAGE_API:-}"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"
RELEASE_ID="${GITHUB_SHA:-$(date -u +%Y%m%d%H%M%S)}"
RELEASE_ID="${RELEASE_ID:0:16}"

log() {
  printf '\033[1;34m==>\033[0m %s\n' "$*"
}

fail() {
  printf '\033[1;31merror:\033[0m %s\n' "$*" >&2
  exit 1
}

command -v ssh >/dev/null || fail "ssh is required"
command -v tar >/dev/null || fail "tar is required"
command -v curl >/dev/null || fail "curl is required"
[[ "${DEPLOY_PATH}" == /* && "${DEPLOY_PATH}" != "/" ]] || fail "DEPLOY_PATH must be an absolute, non-root path"
[[ "${DEPLOY_HOST}" =~ ^[a-zA-Z0-9.-]+$ ]] || fail "DEPLOY_HOST contains unsupported characters"
[[ "${DEPLOY_USER}" =~ ^[a-zA-Z0-9._-]+$ ]] || fail "DEPLOY_USER contains unsupported characters"
[[ "${DEPLOY_DOMAIN}" =~ ^[a-zA-Z0-9.-]+$ ]] || fail "DEPLOY_DOMAIN contains unsupported characters"
[[ "${RELEASE_ID}" =~ ^[a-zA-Z0-9._-]+$ ]] || fail "release ID contains unsupported characters"
[[ -n "${IMAGE_UI}" ]] || fail "IMAGE_UI must be set to the image built by CI"
[[ -n "${IMAGE_API}" ]] || fail "IMAGE_API must be set to the image built by CI"
[[ "${IMAGE_UI}" =~ ^[a-zA-Z0-9./:_-]+$ ]] || fail "IMAGE_UI contains unsupported characters"
[[ "${IMAGE_API}" =~ ^[a-zA-Z0-9./:_-]+$ ]] || fail "IMAGE_API contains unsupported characters"

# ServerAliveInterval keeps the TCP session alive through intermediate NAT while
# a long, silent remote build runs; without it the connection is dropped and ssh
# exits 255 with "client_loop: send disconnect: Broken pipe".
SSH=(ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=30 -o ServerAliveCountMax=20)
if [[ -n "${DEPLOY_SSH_KEY}" ]]; then
  [[ -f "${DEPLOY_SSH_KEY}" ]] || fail "DEPLOY_SSH_KEY does not exist"
  SSH+=(-o IdentitiesOnly=yes -i "${DEPLOY_SSH_KEY}")
fi
SSH+=("${REMOTE}")
REMOTE_RELEASE="${DEPLOY_PATH}/releases/${RELEASE_ID}"

log "Checking SSH access to ${REMOTE}"
"${SSH[@]}" "true" || fail "SSH authentication failed for ${REMOTE}"

log "Checking Docker Compose"
"${SSH[@]}" "docker compose version >/dev/null 2>&1" \
  || fail "Docker Compose is not available on ${REMOTE}"

log "Uploading release ${RELEASE_ID}"
"${SSH[@]}" "mkdir -p '${REMOTE_RELEASE}'"
tar \
  --exclude=.git \
  --exclude=.github \
  --exclude=.keys \
  --exclude=.runs \
  --exclude=.env \
  --exclude='.env.*' \
  --exclude=.ssh \
  --exclude='*.key' \
  --exclude='*.pem' \
  --exclude=github \
  --exclude=github.pub \
  --exclude=node_modules \
  --exclude=.nuxt \
  --exclude=.output \
  -C "${ROOT_DIR}" -cf - . \
  | "${SSH[@]}" "tar -xf - -C '${REMOTE_RELEASE}'"

if [[ -n "${REGISTRY_TOKEN}" ]]; then
  log "Authenticating to ${REGISTRY}"
  # Piped over stdin so the token never lands in the remote process argv.
  printf '%s\n' "${REGISTRY_TOKEN}" \
    | "${SSH[@]}" "docker login '${REGISTRY}' -u '${REGISTRY_USER}' --password-stdin >/dev/null" \
    || fail "docker login to ${REGISTRY} failed"
fi

log "Pulling ${IMAGE_UI} and ${IMAGE_API}"
"${SSH[@]}" bash -s -- \
  "${REMOTE_RELEASE}" "${DEPLOY_PATH}" "${DEPLOY_DOMAIN}" "${IMAGE_UI}" "${IMAGE_API}" <<'REMOTE'
set -Eeuo pipefail
release_path="$1"
deploy_path="$2"
domain="$3"
image_ui="$4"
image_api="$5"

cd "${release_path}/deploy"
export DOMAIN="${domain}" IMAGE_UI="${image_ui}" IMAGE_API="${image_api}"
docker compose pull --quiet api ui
docker compose up -d --remove-orphans
ln -sfn "${release_path}" "${deploy_path}/current"
# Per-commit tags accumulate; drop unused images older than a week so the
# host does not slowly fill up.
docker image prune -af --filter 'until=168h' >/dev/null || true
REMOTE

if [[ -n "${REGISTRY_TOKEN}" ]]; then
  "${SSH[@]}" "docker logout '${REGISTRY}' >/dev/null" || true
fi

log "Waiting for https://${DEPLOY_DOMAIN}/ready"
for attempt in {1..30}; do
  if curl --fail --silent --show-error --max-time 10 "https://${DEPLOY_DOMAIN}/ready" >/dev/null 2>&1; then
    log "Deployment complete: https://${DEPLOY_DOMAIN}"
    exit 0
  fi
  if (( attempt == 30 )); then
    "${SSH[@]}" "cd '${REMOTE_RELEASE}/deploy' && DOMAIN='${DEPLOY_DOMAIN}' IMAGE_UI='${IMAGE_UI}' IMAGE_API='${IMAGE_API}' docker compose ps"
    fail "service did not become ready"
  fi
  sleep 2
done
