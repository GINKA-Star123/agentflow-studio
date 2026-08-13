import json
from enum import StrEnum
from typing import Any

from pydantic import BaseModel,ConfigDict,Field

from app.core.errors import AIRuntimeError
from app.llm.schemas import TokenUsage,ToolCall,ToolCallDelta

class ChatStreamEventType(StrEnum):
    start = "start"
    delta = "delta"
    tool_call_delta = "tool_call_delta"
    usage = "usage"
    done = "done"
    error = "error"

class ChatStreamError(BaseModel):
    code: str
    message: str
    retryable: bool = False
    details: Any | None = None

    model_config = ConfigDict(extra="ignore")

class ChatStreamEvent(BaseModel):
    type: ChatStreamEventType
    delta: str = ""
    text: str = ""
    finish_reason: str | None = None
    token_usage: TokenUsage | None = None
    tool_call_delta: ToolCallDelta | None = None
    tool_calls: list[ToolCall] | None = None
    error: ChatStreamError | None = None
    metadata: dict[str,Any] = Field(default_factory=dict)

    model_config = ConfigDict(extra="ignore")

def sse_encode(event: ChatStreamEvent, request_id:str) -> str:
    payload = event.model_dump(mode="json",exclude_none=True)
    payload["request_id"] = request_id

    return (
        f"event: {event.type.value}\n"
        f"data: {json.dumps(payload, ensure_ascii=False, separators=(',', ':'))}\n\n"
    )

def build_stream_error_event(error:AIRuntimeError) -> ChatStreamEvent:
    return ChatStreamEvent(
        type=ChatStreamEventType.error,
        error=ChatStreamError(
            code=error.code,
            message=error.message,
            retryable=error.retryable,
            details=error.details,
        ),
    )