import logging
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from .llm.factory import create_provider, get_provider_config
from .sql_generator import generate_sql

app = FastAPI(title="QueryForge AI Service", version="1.0.0")
logger = logging.getLogger("queryforge-ai")


class GenerateSQLRequest(BaseModel):
    question: str = Field(min_length=1)
    schema: dict[str, Any]
    safety_rules: list[str] = Field(default_factory=list)


class GenerateSQLResponse(BaseModel):
    sql: str
    explanation: str
    confidence: float


class TestLLMRequest(BaseModel):
    message: str = Field(min_length=1)


@app.on_event("startup")
def startup() -> None:
    config = get_provider_config()
    create_provider(config)
    logger.warning("llm provider configured: provider=%s model=%s base_url=%s", config.provider, config.model, config.base_url)


@app.get("/health")
def health() -> dict[str, str]:
    config = get_provider_config()
    create_provider(config)
    return {"status": "ok", "provider": config.provider, "model": config.model, "base_url": config.base_url}


@app.post("/test-llm")
def test_llm(request: TestLLMRequest) -> dict[str, str]:
    config = get_provider_config()
    provider = create_provider(config)
    messages = [
        {"role": "system", "content": "Reply briefly in plain text."},
        {"role": "user", "content": request.message},
    ]
    try:
        response = provider.generate(messages, temperature=config.temperature)
    except Exception as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return {"provider": config.provider, "model": config.model, "response": response}


@app.post("/generate-sql", response_model=GenerateSQLResponse)
def generate(request: GenerateSQLRequest) -> GenerateSQLResponse:
    try:
        result = generate_sql(request.question, request.schema, request.safety_rules)
    except Exception as exc:
        logger.error("generate-sql failed: %s: %s", type(exc).__name__, exc, exc_info=True)
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    return GenerateSQLResponse(**result)
