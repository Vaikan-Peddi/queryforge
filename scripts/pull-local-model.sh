#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
DEFAULT_MODEL="qwen2.5-coder:7b"

if [[ -f "${ENV_FILE}" ]]; then
  MODEL="$(grep -E '^LLM_MODEL=' "${ENV_FILE}" | tail -n 1 | cut -d '=' -f 2- || true)"
fi
MODEL="${MODEL:-${LLM_MODEL:-${DEFAULT_MODEL}}}"

cd "${ROOT_DIR}"

if command -v docker >/dev/null 2>&1 && docker compose ps --status running --services 2>/dev/null | grep -qx "ollama"; then
  echo "Pulling model into the Docker Compose Ollama service: ${MODEL}"
  docker compose exec -T ollama ollama pull "${MODEL}"
  echo "Model ready in Docker Compose Ollama volume: ${MODEL}"
  exit 0
fi

if ! command -v ollama >/dev/null 2>&1; then
  echo "ollama CLI is not installed."
  echo "Install it first, for example on macOS: brew install ollama"
  echo "Or start the Docker Compose stack first, then rerun this script to pull into the compose ollama service."
  exit 1
fi

if ! ollama list >/dev/null 2>&1; then
  echo "ollama does not appear to be running."
  echo "Start it in another terminal with: ollama serve"
  exit 1
fi

echo "Pulling local model: ${MODEL}"
ollama pull "${MODEL}"
echo "Model ready: ${MODEL}"
