from collections.abc import AsyncIterator

from app.core.errors import ProviderNotFoundError
from app.llm.providers.base import LLMProvider
from app.llm.schemas import ChatRequest, ChatResponse
from app.llm.streaming import ChatStreamEvent


class LLMService:
    def __init__(self) -> None:
        self._providers: dict[str, LLMProvider] = {}

    def register_provider(self, provider: LLMProvider) -> None:
        provider_name = normalize_provider_name(provider.name)
        self._providers[provider_name] = provider

    def get_provider(self, provider_name: str) -> LLMProvider:
        normalized_name = normalize_provider_name(provider_name)
        provider = self._providers.get(normalized_name)

        if provider is None:
            raise ProviderNotFoundError(
                "LLM Provider 不存在",
                details={
                    "provider": provider_name,
                    "available_providers": self.list_provider_names(),
                },
            )

        return provider

    def list_provider_names(self) -> list[str]:
        return sorted(self._providers.keys())

    async def chat(self, request: ChatRequest) -> ChatResponse:
        provider = self.get_provider(request.provider)
        response = await provider.chat(request)
        response.token_usage = response.token_usage.normalized()
        return response

    async def stream_chat(self,request:ChatRequest) -> AsyncIterator[ChatStreamEvent]:
        provider = self.get_provider(request.provider)

        async for event in provider.stream_chat(request):
            if event.token_usage is not None:
                event.token_usage = event.token_usage.normalized()

            yield event


def normalize_provider_name(provider_name: str) -> str:
    return provider_name.strip().lower()