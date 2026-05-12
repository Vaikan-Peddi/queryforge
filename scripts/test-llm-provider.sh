#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

PROVIDER="${LLM_PROVIDER:-ollama}"
BASE_URL="${LLM_BASE_URL:-http://ollama:11434}"
MODEL="${LLM_MODEL:-qwen2.5-coder:7b}"
API_KEY="${LLM_API_KEY:-}"
TEMPERATURE="${LLM_TEMPERATURE:-0.1}"
MESSAGE="${1:-Say hello from QueryForge.}"
JSON_MESSAGE="$(python3 -c 'import json, sys; print(json.dumps(sys.argv[1]))' "${MESSAGE}")"

if [[ "${BASE_URL}" == "http://ollama:"* ]]; then
  BASE_URL="http://localhost:11434"
fi

echo "Testing provider=${PROVIDER} model=${MODEL} base_url=${BASE_URL}"

if [[ "${PROVIDER}" == "ollama" ]]; then
  curl -fsS "${BASE_URL%/}/api/chat" \
    -H "Content-Type: application/json" \
    -d @- <<JSON
{
  "model": "${MODEL}",
  "messages": [{"role": "user", "content": ${JSON_MESSAGE}}],
  "stream": false,
  "options": {"temperature": ${TEMPERATURE}}
}
JSON
  echo
  echo "LLM provider test succeeded."
  exit 0
fi

if [[ "${PROVIDER}" == "openai_compatible" ]]; then
  AUTH_ARGS=()
  if [[ -n "${API_KEY}" ]]; then
    AUTH_ARGS=(-H "Authorization: Bearer ${API_KEY}")
  fi
  curl -fsS "${BASE_URL%/}/chat/completions" \
    -H "Content-Type: application/json" \
    "${AUTH_ARGS[@]}" \
    -d @- <<JSON
{
  "model": "${MODEL}",
  "messages": [{"role": "user", "content": ${JSON_MESSAGE}}],
  "temperature": ${TEMPERATURE}
}
JSON
  echo
  echo "LLM provider test succeeded."
  exit 0
fi

echo "Unsupported LLM_PROVIDER: ${PROVIDER}" >&2
exit 1
