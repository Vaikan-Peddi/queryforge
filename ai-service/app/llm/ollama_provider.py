import httpx

from .base import LLMProvider, ProviderConfig


class OllamaProvider(LLMProvider):
    def __init__(self, config: ProviderConfig):
        self.config = config

    def build_payload(self, messages: list[dict], temperature: float | None = None) -> dict:
        return {
            "model": self.config.model,
            "messages": messages,
            "stream": False,
            "options": {
                "temperature": self.config.temperature if temperature is None else temperature,
            },
        }

    def generate(self, messages: list[dict], temperature: float = 0.1) -> str:
        payload = self.build_payload(messages, temperature)
        with httpx.Client(timeout=self.config.timeout_seconds) as client:
            response = client.post(f"{self.config.base_url.rstrip('/')}/api/chat", json=payload)
            response.raise_for_status()
            data = response.json()
        try:
            return data["message"]["content"]
        except KeyError as exc:
            raise ValueError("Ollama response did not include message.content") from exc
