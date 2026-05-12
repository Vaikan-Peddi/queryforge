import httpx

from .base import LLMProvider, ProviderConfig


class OpenAICompatibleProvider(LLMProvider):
    def __init__(self, config: ProviderConfig):
        self.config = config

    def build_payload(self, messages: list[dict], temperature: float | None = None) -> dict:
        return {
            "model": self.config.model,
            "messages": messages,
            "temperature": self.config.temperature if temperature is None else temperature,
        }

    def build_headers(self) -> dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.config.api_key:
            headers["Authorization"] = f"Bearer {self.config.api_key}"
        return headers

    def generate(self, messages: list[dict], temperature: float = 0.1) -> str:
        payload = self.build_payload(messages, temperature)
        with httpx.Client(timeout=self.config.timeout_seconds) as client:
            response = client.post(
                f"{self.config.base_url.rstrip('/')}/chat/completions",
                headers=self.build_headers(),
                json=payload,
            )
            response.raise_for_status()
            data = response.json()
        try:
            return data["choices"][0]["message"]["content"]
        except (KeyError, IndexError) as exc:
            raise ValueError("OpenAI-compatible response did not include choices[0].message.content") from exc
