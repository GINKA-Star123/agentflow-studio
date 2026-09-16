import asyncio
from collections.abc import AsyncIterator

from app.llm.providers.base import LLMProvider
from app.llm.schemas import ChatMessage, ChatRequest, ChatResponse, ChatRole, TokenUsage
from app.llm.streaming import ChatStreamEvent, ChatStreamEventType


class MockLLMProvider(LLMProvider):
    @property
    def name(self) -> str:
        return "mock"

    async def chat(self, request: ChatRequest) -> ChatResponse:
        user_message = find_latest_user_message(request)
        response_text = build_mock_response_text(request, user_message)

        input_tokens = estimate_tokens(
            "\n".join(message_content_text(message) for message in request.messages)
        )
        output_tokens = estimate_tokens(response_text)

        return ChatResponse(
            text=response_text,
            message=ChatMessage(
                role=ChatRole.assistant,
                content=response_text,
            ),
            token_usage=TokenUsage(
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                total_tokens=input_tokens + output_tokens,
            ),
            raw={
                "provider": self.name,
                "model": request.model,
                "mock": True,
            },
        )

    async def stream_chat(self, request: ChatRequest) -> AsyncIterator[ChatStreamEvent]:
        user_message = find_latest_user_message(request)
        response_text = build_mock_response_text(request, user_message)

        yield ChatStreamEvent(
            type=ChatStreamEventType.start,
            metadata={
                "provider": self.name,
                "model": request.model,
                "mock": True,
            },
        )

        full_text = ""

        for index, chunk in enumerate(split_text(response_text, size=6)):
            await asyncio.sleep(0.03)
            full_text += chunk

            yield ChatStreamEvent(
                type=ChatStreamEventType.delta,
                delta=chunk,
                text=full_text,
                metadata={
                    "provider": self.name,
                    "model": request.model,
                    "index": index,
                    "mock": True,
                },
            )

        input_tokens = estimate_tokens(
            "\n".join(message_content_text(message) for message in request.messages)
        )
        output_tokens = estimate_tokens(response_text)

        token_usage = TokenUsage(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=input_tokens + output_tokens,
        )

        yield ChatStreamEvent(
            type=ChatStreamEventType.usage,
            text=full_text,
            token_usage=token_usage,
            metadata={
                "provider": self.name,
                "model": request.model,
                "mock": True,
            },
        )

        yield ChatStreamEvent(
            type=ChatStreamEventType.done,
            text=full_text,
            finish_reason="stop",
            token_usage=token_usage,
            metadata={
                "provider": self.name,
                "model": request.model,
                "mock": True,
            },
        )


def build_mock_response_text(request: ChatRequest, user_message: ChatMessage) -> str:
    return f"[mock:{request.model}] received request: {message_content_text(user_message)}"


def find_latest_user_message(request: ChatRequest) -> ChatMessage:
    for message in reversed(request.messages):
        if message.role == ChatRole.user:
            return message

    return request.messages[-1]


def estimate_tokens(text: str) -> int:
    stripped = text.strip()
    if not stripped:
        return 0

    return max(1, len(stripped) // 4)


def split_text(text: str, size: int) -> list[str]:
    return [text[index : index + size] for index in range(0, len(text), size)]

def message_content_text(message: ChatMessage) -> str:
    return message.content or ""