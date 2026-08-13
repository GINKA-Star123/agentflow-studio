from typing import Any

from fastapi.encoders import jsonable_encoder
from starlette.requests import Request
from starlette.responses import JSONResponse

def get_request_id(request:Request) -> str:
    return getattr(request.state, "request_id", "")

def ok(request:Request,data:Any) -> dict[str,Any]:
    return {
        "data": jsonable_encoder(data),
        "request_id": get_request_id(request),
    }

def fail(
        request:Request,
        status_code:int,
        code:str,
        message:str,
        details:Any| None = None,
        retryable: bool = False,
) -> JSONResponse:
    payload: dict[str,Any] = {
        "error":{
            "code":code,
            "message":message,
            "retryable":retryable,
        },
        "request_id":get_request_id(request),
    }

    if details is not None:
        payload["error"]["details"] = jsonable_encoder(details)

    return JSONResponse(payload,status_code=status_code)