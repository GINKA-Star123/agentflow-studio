from abc import ABC, abstractmethod
from collections.abc import AsyncIterator

from app.llm.schemas import ChatRequest, ChatResponse
from app.llm.streaming import ChatStreamEvent

class LLMProvider(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        raise NotImplementedError

    @abstractmethod
    async def chat(self, request: ChatRequest) -> ChatResponse:
        raise NotImplementedError

    @abstractmethod
    def stream_chat(self,request:ChatRequest) -> AsyncIterator[ChatStreamEvent]:
        raise NotImplementedError