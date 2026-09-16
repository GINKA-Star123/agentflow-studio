from enum import StrEnum
from typing import Any,Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator,model_validator

class ChatRole(StrEnum):
    system = "system"
    user = "user"
    assistant = "assistant"
    tool = "tool"

def default_tool_parameters() -> dict[str,Any]:
    return {
        "type":"object",
        "properties":{},
    }

class ToolFunctionDefinition(BaseModel):
    name: str = Field(min_length=1, max_length=64, pattern=r"^[a-zA-Z0-9_]+$")
    description: str = ""
    parameters: dict[str,Any] = Field(default_factory=default_tool_parameters)

    model_config = ConfigDict(extra="ignore")

    @field_validator("name","description")
    @classmethod
    def trim_string(cls,value:str) -> str:
        return value.strip()

class ToolDefinition(BaseModel):
    type: Literal["function"] = "function"
    function: ToolFunctionDefinition

    model_config = ConfigDict(extra="ignore")

class ToolChoiceFunction(BaseModel):
    name: str = Field(min_length=1, max_length=64, pattern=r"^[a-zA-Z0-9_]+$")

    model_config = ConfigDict(extra="ignore")

    @field_validator("name")
    @classmethod
    def trim_name(cls,value:str) -> str:
        return value.strip()

class ToolChoiceObject(BaseModel):
    type: Literal["function"] = "function"
    function: ToolChoiceFunction

    model_config = ConfigDict(extra="ignore")

ToolChoice = Literal["none","auto","required"] | ToolChoiceObject

class ToolCallFunction(BaseModel):
    name: str = Field(min_length=1, max_length=64, pattern=r"^[a-zA-Z0-9_]+$")
    arguments: str = ""

    model_config = ConfigDict(extra="ignore")

    @field_validator("name","arguments")
    @classmethod
    def trim_string(cls,value:str) -> str:
        return value.strip()

class ToolCall(BaseModel):
    id: str = Field(min_length=1)
    type: Literal["function"] = "function"
    function: ToolCallFunction

    model_config = ConfigDict(extra="ignore")

    @field_validator("id")
    @classmethod
    def trim_id(cls, value: str) -> str:
        return value.strip()

class ToolCallDelta(BaseModel):
    index: int = Field(ge=0)
    id: str = ""
    type: Literal["function"] = "function"
    function_name: str = ""
    arguments_delta: str = ""

    model_config = ConfigDict(extra="ignore")

class ChatMessage(BaseModel):
    role: ChatRole
    content: str | None = None
    tool_call_id: str | None = None
    tool_calls: list[ToolCall] | None = None

    model_config = ConfigDict(extra="ignore")

    @field_validator("content", "tool_call_id")
    @classmethod
    def trim_optional_string(cls, value: str | None) -> str | None:
        if value is None:
            return None

        return value.strip()

    @model_validator(mode="after")
    def validate_by_role(self) -> "ChatMessage":
        has_content = self.content is not None and self.content != ""
        has_tool_calls = bool(self.tool_calls)
        has_tool_call_id = self.tool_call_id is not None and self.tool_call_id != ""

        if self.role in (ChatRole.system, ChatRole.user):
            if not has_content:
                raise ValueError("system/user message content cannot be empty")

            if has_tool_call_id or has_tool_calls:
                raise ValueError("system/user message cannot include tool fields")

            return self

        if self.role == ChatRole.assistant:
            if has_tool_call_id:
                raise ValueError("assistant message cannot include tool_call_id")

            if not has_content and not has_tool_calls:
                raise ValueError("assistant message requires content or tool_calls")

            return self

        if self.role == ChatRole.tool:
            if not has_tool_call_id:
                raise ValueError("tool message requires tool_call_id")

            if not has_content:
                raise ValueError("tool message content cannot be empty")

            if has_tool_calls:
                raise ValueError("tool message cannot include tool_calls")

            return self

        return self
        

class TokenUsage(BaseModel):
    input_tokens: int = Field(default=0, ge=0)
    output_tokens: int = Field(default=0, ge=0)
    total_tokens: int = Field(default=0, ge=0)

    def normalized(self) -> "TokenUsage":
        if self.total_tokens > 0:
            return self

        return TokenUsage(
            input_tokens=self.input_tokens,
            output_tokens=self.output_tokens,
            total_tokens=self.input_tokens + self.output_tokens
        )

class ChatRequest(BaseModel):
    provider: str = Field(min_length=1)
    model: str = Field(min_length=1)
    messages: list[ChatMessage] = Field(min_length=1)

    temperature: float | None = Field(default=None, ge=0, le=2)
    max_tokens: int | None = Field(default=None, gt=0)
    top_p: float | None = Field(default=None, gt=0,le=1)
    frequency_penalty: float | None = Field(default=None, ge=-2, le=2)
    presence_penalty: float | None = Field(default=None, ge=-2, le=2)
    stop: str | list[str] | None = None

    tools: list[ToolDefinition] = Field(default_factory=list)
    tool_choice: ToolChoice | None = None

    metadata: dict[str,Any] = Field(default_factory=dict)

    model_config = ConfigDict(extra="ignore")

    @field_validator("provider","model")
    @classmethod
    def trim_required_string(cls,value:str) -> str:
        value = value.strip()
        if not value:
            raise ValueError("provider and model cannot be empty")

        return value

    @field_validator("stop")
    @classmethod
    def normalize_stop(cls,value:str|list[str]|None) -> str| list[str] | None:
        if value is None:
            return None

        if isinstance(value,str):
            value=value.strip()
            return value or None

        items = [item.strip() for item in value if item.strip()]
        if len(items)>4:
            raise ValueError("stop 最多支持4个停止序列")

        return items or None

    @model_validator(mode="after")
    def validate_tool_config(self) -> "ChatRequest":
        tool_names = [tool.function.name for tool in self.tools]
        if len(tool_names) != len(set(tool_names)):
            raise ValueError("tool function names must be unique")

        if self.tool_choice is None:
            return self

        if isinstance(self.tool_choice, str):
            if self.tool_choice != "none" and not self.tools:
                raise ValueError("tool_choice requires tools")
            return self

        if not self.tools:
            raise ValueError("tool_choice requires tools")

        selected_name = self.tool_choice.function.name
        if selected_name not in set(tool_names):
            raise ValueError("tool_choice function name must exist in tools")

        return self

class ChatResponse(BaseModel):
    text: str
    message: ChatMessage
    tool_calls: list[ToolCall] = Field(default_factory=list)
    token_usage: TokenUsage = Field(default_factory=TokenUsage)
    raw: dict[str, Any] = Field(default_factory=dict)

    model_config = ConfigDict(extra="ignore")