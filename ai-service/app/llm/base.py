from dataclasses import dataclass


@dataclass(frozen=True)
class ProviderConfig:
    provider: str
    base_url: str
    model: str
    api_key: str = ""
    timeout_seconds: float = 120.0
    temperature: float = 0.1


class LLMProvider:
    config: ProviderConfig

    def generate(self, messages: list[dict], temperature: float = 0.1) -> str:
        raise NotImplementedError
