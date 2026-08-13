from app.core.config import Settings
from app.llm.providers.mock import MockLLMProvider
from app.llm.providers.openai_compatible import OpenAICompatibleProvider
from app.llm.service import LLMService

def create_llm_service(settings: Settings) -> LLMService:
    service = LLMService()

    service.register_provider(MockLLMProvider())

    service.register_provider(
        OpenAICompatibleProvider(
            settings=settings,
            provider_name="openai-compatible",
        )
    )

    service.register_provider(
        OpenAICompatibleProvider(
            settings=settings,
            provider_name="openai",
        )
    )

    return service