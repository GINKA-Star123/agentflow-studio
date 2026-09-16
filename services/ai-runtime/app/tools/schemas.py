import json
from enum import StrEnum
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator,field_validator

class ToolCallStatus(StrEnum):
    pending = "pending"
    running = "running"
    succeeded = "succeeded"
    failed = "failed"

class ToolCallError(BaseModel):
    code: str
    message: str
    retryable: bool = False
    details: Any | None = None

    model_config = ConfigDict(extra="ignore")

class ToolCallRequest(BaseModel):
    tool_call_id: str = Field(min_length=1)
    tool_name: str = Field(min_length=1, max_length=64, pattern=r"^[a-zA-Z0-9_]+$")
    arguments: dict[str,Any] = Field(default_factory=dict)

    workspace_id: str = ""
    workflow_id: str = ""
    run_id: str = ""
    node_id: str = ""

    timeout_ms: int = Field(default=30000,gt=0,le=300000)
    metadata: dict[str,Any] = Field(default_factory=dict)

    model_config = ConfigDict(extra="ignore")

    @field_validator("tool_call_id", "tool_name", "workspace_id", "workflow_id", "run_id", "node_id")
    @classmethod
    def trim_string(cls, value: str) -> str:
        return value.strip()

class ToolCallResponse(BaseModel):
    tool_call_id: str = Field(min_length=1)
    tool_name: str = Field(min_length=1)
    status: ToolCallStatus
    result: Any | None = None
    error: ToolCallError | None = None
    latency_ms: int = Field(default=0, ge=0)
    metadata: dict[str, Any] = Field(default_factory=dict)

    model_config = ConfigDict(extra="ignore")

    @field_validator("tool_call_id", "tool_name")
    @classmethod
    def trim_string(cls, value: str) -> str:
        return value.strip()

    @model_validator(mode="after")
    def validate_status_payload(self) -> "ToolCallResponse":
        if self.status == ToolCallStatus.succeeded and self.error is not None:
            raise ValueError("succeeded tool call cannot include error")

        if self.status == ToolCallStatus.failed and self.error is None:
            raise ValueError("failed tool call requires error")

        return self


class ToolResultMessage(BaseModel):
    role: Literal["tool"] = "tool"
    tool_call_id: str = Field(min_length=1)
    content: str = Field(min_length=1)

    model_config = ConfigDict(extra="ignore")


def build_tool_result_message(response: ToolCallResponse) -> ToolResultMessage:
    payload = {
        "status": response.status.value,
        "result": response.result,
        "error": response.error.model_dump(mode="json", exclude_none=True) if response.error else None,
        "metadata": response.metadata,
    }

    return ToolResultMessage(
        tool_call_id=response.tool_call_id,
        content=json.dumps(payload, ensure_ascii=False, separators=(",", ":")),
    )