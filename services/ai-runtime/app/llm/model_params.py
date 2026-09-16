from dataclasses import dataclass
from typing import Any

from app.core.config import Settings
from app.llm.schemas import ChatRequest

@dataclass(frozen=True)
class ModelParameters:
    temperature: float
    max_tokens: int
    top_p: float | None = None
    frequency_penalty: float | None = None
    presence_penalty: float | None = None
    stop:str|list[str]|None = None

    def to_openai_options(self) -> dict[str,Any]:
        options :dict[str,Any] = {
            "temperature":self.temperature,
            "max_tokens":self.max_tokens,
        }

        if self.top_p is not None:
            options["top_p"] = self.top_p

        if self.frequency_penalty is not None:
            options["frequency_penalty"] = self.frequency_penalty

        if self.presence_penalty is not None:
            options["presence_penalty"] = self.presence_penalty

        if self.stop is not None:
            options["stop"] = self.stop

        return options
 
    def to_public_dict(self) -> dict[str,Any]:
        return self.to_openai_options()


def normalize_model_parameters(request:ChatRequest,settings:Settings) -> ModelParameters:
    temperature = (
        request.temperature
        if request.temperature is not None
        else settings.llm_default_temperature
    )

    max_tokens = (
        request.max_tokens
        if request.max_tokens is not None
        else settings.llm_default_max_tokens
    )

    max_tokens = min(max_tokens, settings.llm_max_tokens_limit)

    return ModelParameters(
        temperature=temperature,
        max_tokens=max_tokens,
        top_p=request.top_p,
        frequency_penalty=request.frequency_penalty,
        presence_penalty=request.presence_penalty,
        stop=request.stop
    )