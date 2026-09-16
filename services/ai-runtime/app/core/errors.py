from typing import Any

class AIRuntimeError(Exception):
    code = "AI_RUNTIME_ERROR"
    status_code = 500
    retryable = False

    def __init__(
            self,
            message:str,
            *,
            code:str|None =None,
            status_code:int|None = None,
            details:Any|None = None,
            retryable: bool | None = None,
    ) -> None:
        super().__init__(message)

        if code is not None:
            self.code = code

        if status_code is not None:
            self.status_code = status_code

        if retryable is not None:
            self.retryable = retryable
        else:
            self.retryable = bool(getattr(self,"retryable",False))

        self.message = message
        self.details = details

class InvalidRequestError(AIRuntimeError):
    code = "AI_RUNTIME_INVALID_REQUEST"
    status_code = 400
    retryable = False


class ProviderNotFoundError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_NOT_FOUND"
    status_code = 400
    retryable = False


class ProviderConfigError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_CONFIG_INVALID"
    status_code = 500
    retryable = False


class ProviderAuthError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_AUTH_FAILED"
    status_code = 502
    retryable = False


class ProviderRateLimitError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_RATE_LIMIT"
    status_code = 429
    retryable = True


class ProviderTimeoutError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_TIMEOUT"
    status_code = 504
    retryable = True


class ProviderCallError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_CALL_FAILED"
    status_code = 502
    retryable = False


class ProviderResponseError(AIRuntimeError):
    code = "AI_RUNTIME_PROVIDER_RESPONSE_INVALID"
    status_code = 502
    retryable = False