from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings,SettingsConfigDict

class Settings(BaseSettings):
    app_env :str = "dev"
    ai_runtime_http_port: int = 8090

    openai_compatible_base_url:str = "https://api.openai.com/v1"
    openai_compatible_api_key:str = "change-me"
    openai_compatible_stream_include_usage: bool = False

    llm_default_provider: str = "mock"
    llm_default_model: str = "mock-chat"

    llm_request_timeout_seconds:float = Field(default=60.0,gt=0.0)
    llm_max_retries: int = Field(default=2, ge=0, le=5)
    llm_retry_initial_delay_seconds: float = Field(default=0.5, ge=0)
    llm_retry_max_delay_seconds: float = Field(default=4.0, ge=0)

    llm_default_temperature: float = Field(default=0.7, ge=0,le=2)
    llm_default_max_tokens: int = Field(default=1024,gt=0)
    llm_max_tokens_limit: int = Field(default=4096,gt=0)

    qdrant_url: str = "http://localhost:6333"
    qdrant_api_key:str = ""

    otel_exporter_otlp_endpoint: str = "http://localhost:4318"

    model_config = SettingsConfigDict(
        env_file = ("../../.env", ".env"),
        env_file_encoding = "utf-8",
        extra="ignore",
    )


@lru_cache
def get_settings() -> Settings:
    return Settings()