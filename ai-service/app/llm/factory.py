import os

from .base import LLMProvider, ProviderConfig
from .ollama_provider import OllamaProvider
from .openai_compatible_provider import OpenAICompatibleProvider


def get_provider_config() -> ProviderConfig:
    return ProviderConfig(
        provider=os.getenv("LLM_PROVIDER", "ollama").strip().lower(),
        base_url=os.getenv("LLM_BASE_URL", "http://ollama:11434").strip().rstrip("/"),
        model=os.getenv("LLM_MODEL", "qwen2.5-coder:7b").strip(),
        api_key=os.getenv("LLM_API_KEY", "").strip(),
        timeout_seconds=_float_env("LLM_TIMEOUT_SECONDS", 120.0),
        temperature=_float_env("LLM_TEMPERATURE", 0.1),
    )


def create_provider(config: ProviderConfig | None = None) -> LLMProvider:
    config = config or get_provider_config()
    if config.provider == "ollama":
        return OllamaProvider(config)
    if config.provider == "openai_compatible":
        return OpenAICompatibleProvider(config)
    raise ValueError("Unsupported LLM_PROVIDER %r. Supported values: ollama, openai_compatible" % config.provider)


def _float_env(name: str, default: float) -> float:
    value = os.getenv(name, "")
    if not value:
        return default
    try:
        return float(value)
    except ValueError as exc:
        raise ValueError(f"{name} must be a number") from exc
