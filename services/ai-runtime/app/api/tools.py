from fastapi import APIRouter
from starlette.requests import Request

from app.api.response import fail
from app.tools.schemas import ToolCallRequest

router = APIRouter(prefix="/tools", tags=["tools"])

@router.post("/call")
async def call_tool(request: Request, payload: ToolCallRequest):
    return fail(
        request=request,
        status_code=501,
        code="AI_RUNTIME_TOOL_EXECUTION_NOT_IMPLEMENTED",
        message="AI Runtime tool calling protocol is defined,but tool execution not implemented",
        details={
            "tool_name": payload.tool_name,
            "tool_call_id": payload.tool_call_id,
            "protocol_version": "1.0.0",
        },
        retryable=False,
    )