import pytest

from app.llm.base import ProviderConfig
from app.llm.factory import create_provider, get_provider_config
from app.llm.ollama_provider import OllamaProvider
from app.llm.openai_compatible_provider import OpenAICompatibleProvider
from app.sql_generator import parse_model_json


def test_provider_factory_defaults_to_ollama(monkeypatch):
    for name in ("LLM_PROVIDER", "LLM_BASE_URL", "LLM_MODEL", "LLM_API_KEY", "LLM_TIMEOUT_SECONDS", "LLM_TEMPERATURE"):
        monkeypatch.delenv(name, raising=False)

    config = get_provider_config()
    provider = create_provider(config)

    assert config.provider == "ollama"
    assert config.base_url == "http://ollama:11434"
    assert config.model == "qwen2.5-coder:7b"
    assert isinstance(provider, OllamaProvider)


def test_provider_factory_rejects_unknown_provider():
    config = ProviderConfig(provider="mystery", base_url="http://example.test", model="model")

    with pytest.raises(ValueError, match="Unsupported LLM_PROVIDER"):
        create_provider(config)


def test_ollama_payload_formatting():
    provider = OllamaProvider(
        ProviderConfig(provider="ollama", base_url="http://ollama:11434", model="qwen2.5-coder:7b", temperature=0.2)
    )
    messages = [{"role": "user", "content": "hello"}]

    payload = provider.build_payload(messages)

    assert payload == {
        "model": "qwen2.5-coder:7b",
        "messages": messages,
        "stream": False,
        "options": {"temperature": 0.2},
    }


def test_openai_compatible_payload_and_headers_with_key():
    provider = OpenAICompatibleProvider(
        ProviderConfig(
            provider="openai_compatible",
            base_url="https://api.example.test/v1",
            model="provider-model",
            api_key="secret",
            temperature=0.3,
        )
    )
    messages = [{"role": "user", "content": "hello"}]

    assert provider.build_payload(messages) == {
        "model": "provider-model",
        "messages": messages,
        "temperature": 0.3,
    }
    assert provider.build_headers()["Authorization"] == "Bearer secret"


def test_openai_compatible_omits_authorization_without_key():
    provider = OpenAICompatibleProvider(
        ProviderConfig(provider="openai_compatible", base_url="http://localhost:1234/v1", model="local-model")
    )

    assert "Authorization" not in provider.build_headers()


def test_json_extraction_from_messy_model_response():
    content = """
    Sure, here is the JSON:
    {"sql":"SELECT name FROM customers LIMIT 10","explanation":"Lists customers.","confidence":0.81}
    Thanks.
    """

    parsed = parse_model_json(content)

    assert parsed["sql"] == "SELECT name FROM customers LIMIT 10"
    assert parsed["confidence"] == 0.81


def test_json_extraction_failure():
    with pytest.raises(ValueError, match="valid JSON"):
        parse_model_json("I refuse to return JSON")
