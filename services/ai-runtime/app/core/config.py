from functools import lru_cache

from pydantic_settings import BaseSettings,SettingsConfigDict

class Settings(BaseSettings):
    app_env :str = "dev"
    ai_runtime_http_port: int = 8090

    openai_compatible_base_url:str = "https://api.openai.com/v1"
    oepnai_compatible_api_key:str = "change-me"

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