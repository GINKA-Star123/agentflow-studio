import asyncio
import logging
import time
from typing import Any
from collections.abc import AsyncIterator

from openai import APIConnectionError, APIStatusError, APITimeoutError, AsyncOpenAI, OpenAIError

from app.core.config import Settings
from app.core.errors import (
    InvalidRequestError,
    ProviderAuthError,
    ProviderCallError,
    ProviderConfigError,
    ProviderRateLimitError,
    ProviderResponseError,
    ProviderTimeoutError,
    AIRuntimeError,
)
from app.llm.model_params import ModelParameters, normalize_model_parameters
from app.llm.providers.base import LLMProvider
from app.llm.schemas import (
    ChatMessage,
    ChatRequest,
    ChatResponse,
    ChatRole,
    TokenUsage,
    ToolCall,
    ToolCallDelta,
)
from app.llm.streaming import ChatStreamEventType, ChatStreamEvent

logger = logging.getLogger("ai_runtime.llm.openai_compatible")


class OpenAICompatibleProvider(LLMProvider):
    def __init__(self, settings: Settings, provider_name: str = "openai-compatible") -> None:
        self._settings = settings
        self._name = provider_name
        self._base_url = settings.openai_compatible_base_url.rstrip("/")
        self._api_key = settings.openai_compatible_api_key
        self._timeout = settings.llm_request_timeout_seconds
        self._max_retries = settings.llm_max_retries
        self._retry_initial_delay = settings.llm_retry_initial_delay_seconds
        self._retry_max_delay = settings.llm_retry_max_delay_seconds
        self._stream_include_usage = settings.openai_compatible_stream_include_usage

        self._client = AsyncOpenAI(
            base_url=self._base_url,
            api_key=self._api_key,
            timeout=self._timeout,
        )

    @property
    def name(self) -> str:
        return self._name

    async def chat(self, request: ChatRequest) -> ChatResponse:
        self._ensure_configured()

        messages = build_openai_messages(request)
        parameters = normalize_model_parameters(request, self._settings)
        max_attempts = max(1, self._max_retries + 1)

        for attempt in range(1, max_attempts + 1):
            started_at = time.perf_counter()

            try:
                options = parameters.to_openai_options()
                options.update(build_openai_tool_options(request))
                
                response = await self._client.chat.completions.create( # type: ignore
                    model=request.model,
                    messages=messages, # type: ignore
                    **options,
                )

                latency_ms = int((time.perf_counter() - started_at) * 1000)

                return build_chat_response(
                    response=response,
                    request=request,
                    provider_name=self.name,
                    base_url=self._base_url,
                    parameters=parameters,
                    attempt=attempt,
                    latency_ms=latency_ms,
                )

            except (APIStatusError, APITimeoutError, APIConnectionError, OpenAIError) as error:
                runtime_error = normalize_openai_error(
                    error=error,
                    request=request,
                    provider_name=self.name,
                    base_url=self._base_url,
                    timeout_seconds=self._timeout,
                    attempt=attempt,
                    max_attempts=max_attempts,
                )

                if not should_retry_openai_error(error) or attempt >= max_attempts:
                    raise runtime_error from error

                delay_seconds = self._retry_delay_seconds(attempt)

                logger.warning(
                    "llm provider request failed, retrying",
                    extra={
                        "provider": self.name,
                        "model": request.model,
                        "attempt": attempt,
                        "max_attempts": max_attempts,
                        "delay_seconds": delay_seconds,
                        "error_code": runtime_error.code,
                    },
                )

                await asyncio.sleep(delay_seconds)

        raise ProviderCallError(
            "OpenAI Compatible Provider 调用失败",
            retryable=True,
            details={
                "provider": self.name,
                "model": request.model,
                "base_url": self._base_url,
                "max_attempts": max_attempts,
            },
        )

    async def stream_chat(self, request: ChatRequest) -> AsyncIterator[ChatStreamEvent]:
        self._ensure_configured()

        messages = build_openai_messages(request)
        parameters = normalize_model_parameters(request, self._settings)

        started_at = time.perf_counter()
        stream, attempt = await self._create_stream_with_retry(
            request=request,
            messages=messages,
            parameters=parameters,
        )

        yield ChatStreamEvent(
            type=ChatStreamEventType.start,
            metadata={
                "provider": self.name,
                "model": request.model,
                "base_url": self._base_url,
                "attempt": attempt,
                "stream": True,
            },
        )

        text_parts: list[str] = []
        token_usage = TokenUsage()
        finish_reason: str | None = None
        tool_call_state: dict[int, dict[str, Any]] = {}

        try:
            async for chunk in stream:
                usage = getattr(chunk, "usage", None)
                if usage is not None:
                    token_usage = token_usage_from_openai_usage(usage)

                    yield ChatStreamEvent(
                        type=ChatStreamEventType.usage,
                        text="".join(text_parts),
                        token_usage=token_usage,
                        metadata={
                            "provider": self.name,
                            "model": request.model,
                            "base_url": self._base_url,
                            "attempt": attempt,
                        },
                    )

                    continue

                choices = getattr(chunk, "choices", None) or []
                if not choices:
                    continue

                choice = choices[0]
                delta = getattr(choice, "delta", None)
                if delta is None:
                    continue

                content = getattr(delta, "content", None)
                if content:
                    text_parts.append(content)

                    yield ChatStreamEvent(
                        type=ChatStreamEventType.delta,
                        delta=content,
                        text="".join(text_parts),
                        metadata={
                            "provider": self.name,
                            "model": request.model,
                            "base_url": self._base_url,
                            "attempt": attempt,
                        },
                    )

                raw_tool_call_deltas = getattr(delta, "tool_calls", None) or []
                for raw_tool_call_delta in raw_tool_call_deltas:
                    parsed_delta = parse_openai_tool_call_delta(raw_tool_call_delta)
                    if parsed_delta is None:
                        continue

                    apply_tool_call_delta(tool_call_state, parsed_delta)

                    yield ChatStreamEvent(
                        type=ChatStreamEventType.tool_call_delta,
                        text="".join(text_parts),
                        tool_call_delta=parsed_delta,
                        metadata={
                            "provider": self.name,
                            "model": request.model,
                            "base_url": self._base_url,
                            "attempt": attempt,
                        },
                    )

                choice_finish_reason = getattr(choice, "finish_reason", None)
                if choice_finish_reason:
                    finish_reason = choice_finish_reason

        except (APIStatusError, APITimeoutError, APIConnectionError, OpenAIError) as error:
            runtime_error = normalize_openai_error(
                error=error,
                request=request,
                provider_name=self.name,
                base_url=self._base_url,
                timeout_seconds=self._timeout,
                attempt=attempt,
                max_attempts=attempt,
            )

            raise ProviderCallError(
                "OpenAI Compatible Provider stream request failed",
                retryable=False,
                details={
                    "provider": self.name,
                    "model": request.model,
                    "base_url": self._base_url,
                    "partial_text_length": len("".join(text_parts)),
                    "reason": runtime_error.message,
                    "error_code": runtime_error.code,
                    "error_details": runtime_error.details,
                },
            ) from error

        latency_ms = int((time.perf_counter() - started_at) * 1000)
        full_text = "".join(text_parts)
        tool_calls = build_tool_calls_from_stream_state(tool_call_state)

        yield ChatStreamEvent(
            type=ChatStreamEventType.done,
            text=full_text,
            finish_reason=finish_reason or "stop",
            token_usage=token_usage,
            tool_calls=tool_calls or None,
            metadata={
                "provider": self.name,
                "model": request.model,
                "base_url": self._base_url,
                "attempt": attempt,
                "latency_ms": latency_ms,
                "parameters": parameters.to_public_dict(),
            },
        )

    async def _create_stream_with_retry(
            self,
            *,
            request: ChatRequest,
            messages: list[dict[str,Any]],
            parameters: ModelParameters,
    ) -> tuple[Any,int]:
        max_attempts = max(1,self._max_retries+1)

        for attempt in range(1,max_attempts+1):
            try:
                options = parameters.to_openai_options()
                options.update(build_openai_tool_options(request))
                options["stream"] = True

                if self._stream_include_usage:
                    options["stream_options"] = {
                        "include_usage": True,
                    }

                stream = await self._client.chat.completions.create( # type: ignore
                    model=request.model,
                    messages=messages, # type: ignore
                    **options,
                )

                return stream,attempt
            except (APIStatusError,APITimeoutError,APIConnectionError,OpenAIError) as error:
                runtime_error = normalize_openai_error(
                    error=error,
                    request=request,
                    provider_name=self.name,
                    base_url=self._base_url,
                    timeout_seconds=self._timeout,
                    attempt=attempt,
                    max_attempts=max_attempts,
                )

                if not should_retry_openai_error(error) or attempt >= max_attempts:
                    raise runtime_error from error

                delay_seconds = self._retry_delay_seconds(attempt)

                logger.warning(
                "llm provider stream request failed, retrying",
                extra={
                    "provider": self.name,
                    "model": request.model,
                    "attempt": attempt,
                    "max_attempts": max_attempts,
                    "delay_seconds": delay_seconds,
                    "error_code": runtime_error.code,
                },
            )

            await asyncio.sleep(delay_seconds)

        raise ProviderCallError(
            "OpenAI Compatible Provider 调用失败",
            retryable=False,
            details={
                "providers":self.name,
                "model":request.model,
                "base_url":self._base_url,
                "max_attempts":max_attempts,
            },
        )

    def _ensure_configured(self) -> None:
        if not self._base_url:
            raise ProviderConfigError(
                "OPENAI_COMPATIBLE_BASE_URL 未配置",
                details={
                    "provider": self.name,
                    "env": "OPENAI_COMPATIBLE_BASE_URL",
                },
            )

        if not self._api_key or self._api_key == "change-me":
            raise ProviderConfigError(
                "OPENAI_COMPATIBLE_API_KEY 未配置",
                details={
                    "provider": self.name,
                    "base_url": self._base_url,
                    "env": "OPENAI_COMPATIBLE_API_KEY",
                },
            )

    def _retry_delay_seconds(self, attempt: int) -> float:
        delay = self._retry_initial_delay * (2 ** (attempt - 1))
        return min(delay, self._retry_max_delay)


def build_openai_messages(request: ChatRequest) -> list[dict[str, Any]]:
    messages: list[dict[str, Any]] = []

    for message in request.messages:
        payload: dict[str, Any] = {
            "role": message.role.value,
        }

        if message.content is not None:
            payload["content"] = message.content

        if message.role == ChatRole.tool:
            payload["tool_call_id"] = message.tool_call_id

        if message.tool_calls:
            payload["tool_calls"] = [
                tool_call_to_openai_dict(tool_call)
                for tool_call in message.tool_calls
            ]

        messages.append(payload)

    return messages

def build_chat_response(
    *,
    response: Any,
    request: ChatRequest,
    provider_name: str,
    base_url: str,
    parameters: ModelParameters,
    attempt: int,
    latency_ms: int,
) -> ChatResponse:
    response_text = extract_response_text(
        response=response,
        provider_name=provider_name,
        model=request.model,
    )

    tool_calls = extract_tool_calls(response)
    token_usage = extract_token_usage(response)

    return ChatResponse(
        text=response_text,
        message=ChatMessage(
            role=ChatRole.assistant,
            content=response_text,
            tool_calls=tool_calls or None,
        ),
        tool_calls=tool_calls,
        token_usage=token_usage,
        raw={
            "provider": provider_name,
            "model": request.model,
            "base_url": base_url,
            "openai_response_id": getattr(response, "id", ""),
            "finish_reason": extract_finish_reason(response),
            "attempts": attempt,
            "latency_ms": latency_ms,
            "parameters": parameters.to_public_dict(),
        },
    )

def extract_response_text(*, response: Any, provider_name: str, model: str) -> str:
    choices = getattr(response, "choices", None) or []
    if not choices:
        raise ProviderResponseError(
            "OpenAI Compatible Provider 返回结果缺少 choices",
            details={
                "provider": provider_name,
                "model": model,
                "openai_response_id": getattr(response, "id", ""),
            },
        )

    message = choices[0].message
    content = message.content

    if content is None:
        return ""

    if isinstance(content, str):
        return content

    return str(content)


def extract_token_usage(response: Any) -> TokenUsage:
    usage = getattr(response, "usage", None)
    if usage is None:
        return TokenUsage()

    input_tokens = getattr(usage, "prompt_tokens", 0) or 0
    output_tokens = getattr(usage, "completion_tokens", 0) or 0
    total_tokens = getattr(usage, "total_tokens", 0) or 0

    return TokenUsage(
        input_tokens=input_tokens,
        output_tokens=output_tokens,
        total_tokens=total_tokens,
    ).normalized()


def normalize_openai_error(
    *,
    error: OpenAIError,
    request: ChatRequest,
    provider_name: str,
    base_url: str,
    timeout_seconds: float,
    attempt: int,
    max_attempts: int,
) -> ProviderCallError:
    details: dict[str, Any] = {
        "provider": provider_name,
        "model": request.model,
        "base_url": base_url,
        "attempt": attempt,
        "max_attempts": max_attempts,
    }

    if isinstance(error, APIStatusError):
        details["provider_status_code"] = error.status_code
        details["provider_response"] = safe_response_text(error)

        if error.status_code in (401, 403):
            return ProviderAuthError( # type: ignore
                "OpenAI Compatible Provider 鉴权失败",
                details=details,
            )

        if error.status_code == 429:
            return ProviderRateLimitError( # type: ignore
                "OpenAI Compatible Provider 请求被限流",
                details=details,
            )

        if error.status_code in (408, 504):
            return ProviderTimeoutError( # type: ignore
                "OpenAI Compatible Provider 返回超时状态",
                details=details,
            )

        return ProviderCallError(
            "OpenAI Compatible Provider 返回错误状态码",
            retryable=is_retryable_status_code(error.status_code),
            details=details,
        )

    if isinstance(error, APITimeoutError):
        details["timeout_seconds"] = timeout_seconds
        return ProviderTimeoutError( # type: ignore
            "OpenAI Compatible Provider 请求超时",
            details=details,
        )

    if isinstance(error, APIConnectionError):
        details["reason"] = str(error)
        return ProviderCallError(
            "OpenAI Compatible Provider 连接失败",
            retryable=True,
            details=details,
        )

    details["reason"] = str(error)
    return ProviderCallError(
        "OpenAI Compatible Provider 调用失败",
        retryable=False,
        details=details,
    )


def should_retry_openai_error(error: OpenAIError) -> bool:
    if isinstance(error, (APITimeoutError, APIConnectionError)):
        return True

    if isinstance(error, APIStatusError):
        return is_retryable_status_code(error.status_code)

    return False


def is_retryable_status_code(status_code: int) -> bool:
    return status_code in (408, 409, 429) or status_code >= 500


def safe_response_text(error: APIStatusError) -> str:
    response = getattr(error, "response", None)
    if response is None:
        return ""

    text = getattr(response, "text", "")
    if not text:
        return ""

    return text[:1000]

def token_usage_from_openai_usage(usage:Any) -> TokenUsage:
    input_tokens = getattr(usage,"prompt_tokens",0) or 0
    output_tokens = getattr(usage,"completion_tokens",0) or 0
    total_tokens = getattr(usage,"total_tokens",0) or 0

    return TokenUsage(
        input_tokens=input_tokens,
        output_tokens=output_tokens,
        total_tokens=total_tokens
    ).normalized()

def build_openai_tool_options(request: ChatRequest) -> dict[str, Any]:
    options: dict[str, Any] = {}

    if request.tools:
        options["tools"] = [
            tool.model_dump(mode="json", exclude_none=True)
            for tool in request.tools
        ]

    if request.tool_choice is not None:
        if isinstance(request.tool_choice, str):
            options["tool_choice"] = request.tool_choice
        else:
            options["tool_choice"] = request.tool_choice.model_dump(
                mode="json",
                exclude_none=True,
            )

    return options


def tool_call_to_openai_dict(tool_call: ToolCall) -> dict[str, Any]:
    return {
        "id": tool_call.id,
        "type": tool_call.type,
        "function": {
            "name": tool_call.function.name,
            "arguments": tool_call.function.arguments,
        },
    }


def extract_tool_calls(response: Any) -> list[ToolCall]:
    choices = getattr(response, "choices", None) or []
    if not choices:
        return []

    message = getattr(choices[0], "message", None)
    if message is None:
        return []

    return extract_tool_calls_from_message(message)


def extract_tool_calls_from_message(message: Any) -> list[ToolCall]:
    raw_tool_calls = getattr(message, "tool_calls", None) or []
    tool_calls: list[ToolCall] = []

    for raw_tool_call in raw_tool_calls:
        function = getattr(raw_tool_call, "function", None)
        if function is None:
            continue

        tool_call_id = getattr(raw_tool_call, "id", "") or ""
        function_name = getattr(function, "name", "") or ""
        arguments = getattr(function, "arguments", "") or ""

        if not tool_call_id or not function_name:
            continue

        tool_calls.append(
            ToolCall(
                id=tool_call_id,
                type="function",
                function={ # type: ignore
                    "name": function_name,
                    "arguments": arguments,
                },
            )
        )

    return tool_calls


def extract_finish_reason(response: Any) -> str:
    choices = getattr(response, "choices", None) or []
    if not choices:
        return ""

    return getattr(choices[0], "finish_reason", "") or ""


def parse_openai_tool_call_delta(raw_delta: Any) -> ToolCallDelta | None:
    index = getattr(raw_delta, "index", None)
    if index is None:
        return None

    function = getattr(raw_delta, "function", None)
    function_name = ""
    arguments_delta = ""

    if function is not None:
        function_name = getattr(function, "name", "") or ""
        arguments_delta = getattr(function, "arguments", "") or ""

    tool_call_id = getattr(raw_delta, "id", "") or ""
    tool_call_type = getattr(raw_delta, "type", "") or "function"

    if tool_call_type != "function":
        return None

    if not tool_call_id and not function_name and not arguments_delta:
        return None

    return ToolCallDelta(
        index=index,
        id=tool_call_id,
        type="function",
        function_name=function_name,
        arguments_delta=arguments_delta,
    )


def apply_tool_call_delta(
    state: dict[int, dict[str, Any]],
    delta: ToolCallDelta,
) -> None:
    entry = state.setdefault(
        delta.index,
        {
            "id": "",
            "type": "function",
            "function": {
                "name": "",
                "arguments": "",
            },
        },
    )

    if delta.id:
        entry["id"] = delta.id

    function_state = entry.setdefault(
        "function",
        {
            "name": "",
            "arguments": "",
        },
    )

    if delta.function_name:
        function_state["name"] = delta.function_name

    if delta.arguments_delta:
        function_state["arguments"] = function_state.get("arguments", "") + delta.arguments_delta


def build_tool_calls_from_stream_state(
    state: dict[int, dict[str, Any]],
) -> list[ToolCall]:
    tool_calls: list[ToolCall] = []

    for index in sorted(state.keys()):
        entry = state[index]
        function_state = entry.get("function", {})

        tool_call_id = entry.get("id", "")
        function_name = function_state.get("name", "")
        arguments = function_state.get("arguments", "")

        if not tool_call_id or not function_name:
            continue

        tool_calls.append(
            ToolCall(
                id=tool_call_id,
                type="function",
                function={ # type: ignore
                    "name": function_name,
                    "arguments": arguments,
                },
            )
        )

    return tool_calls