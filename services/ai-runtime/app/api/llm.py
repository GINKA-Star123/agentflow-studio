import asyncio
import logging

from fastapi import APIRouter
from starlette.requests import Request
from starlette.responses import StreamingResponse

from app.api.response import fail,ok
from app.core.errors import AIRuntimeError
from app.llm.schemas import ChatRequest
from app.llm.service import LLMService
from app.llm.streaming import build_stream_error_event, sse_encode

router = APIRouter(prefix="/llm",tags=["llm"])
logger = logging.getLogger("ai_runtime.llm")

@router.post("/chat")
async def chat(request: Request, payload: ChatRequest):
    service = get_llm_service(request)

    try:
        response = await service.chat(payload)

    except AIRuntimeError as error:
        logger.warning(
            "llm chat 失败",
            extra={
                "request_id": getattr(request.state, "request_id", ""),
                "trace_id": payload.metadata.get("trace_id", ""),
                "run_id": payload.metadata.get("run_id", ""),
                "node_id": payload.metadata.get("node_id", ""),
                "error_code":error.code,
                "retryable":error.retryable,
            },
        )
        return fail(
            request,
            error.status_code,
            error.code,
            error.message,
            error.details,
            error.retryable,
        )

    except Exception as error:
        logger.exception(
            "llm chat 未处理异常",
            extra={
                "request_id": getattr(request.state, "request_id", ""),
                "trace_id": payload.metadata.get("trace_id", ""),
                "run_id": payload.metadata.get("run_id", ""),
                "node_id": payload.metadata.get("node_id", ""),
            },
        )
        return fail(
            request,
            500,
            "AI_RUNTIME_INTERNAL",
            "AI Runtime 内部错误",
            {
                "reason": str(error),
            },
            retryable=False,
        )

    return ok(request, response)

@router.post("/stream")
async def stream_chat(request: Request, payload: ChatRequest):
    service = get_llm_service(request)
    request_id = getattr(request.state, "request_id", "")

    async def event_generator():
        try:
            async for event in service.stream_chat(payload):
                if await request.is_disconnected():
                    logger.info(
                        "llm stream client disconnected",
                        extra={
                            "request_id": request_id,
                            "trace_id": payload.metadata.get("trace_id", ""),
                            "run_id": payload.metadata.get("run_id", ""),
                            "node_id": payload.metadata.get("node_id", ""),
                        },
                    )
                    break

                yield sse_encode(event, request_id)

        except asyncio.CancelledError:
            logger.info(
                "llm stream cancelled",
                extra={
                    "request_id": request_id,
                    "trace_id": payload.metadata.get("trace_id", ""),
                    "run_id": payload.metadata.get("run_id", ""),
                    "node_id": payload.metadata.get("node_id", ""),
                },
            )
            raise

        except AIRuntimeError as error:
            logger.warning(
                "llm stream failed",
                extra={
                    "request_id": request_id,
                    "trace_id": payload.metadata.get("trace_id", ""),
                    "run_id": payload.metadata.get("run_id", ""),
                    "node_id": payload.metadata.get("node_id", ""),
                    "error_code": error.code,
                    "retryable": error.retryable,
                },
            )

            yield sse_encode(build_stream_error_event(error), request_id)

        except Exception as error:
            logger.exception(
                "llm stream unexpected error",
                extra={
                    "request_id": request_id,
                    "trace_id": payload.metadata.get("trace_id", ""),
                    "run_id": payload.metadata.get("run_id", ""),
                    "node_id": payload.metadata.get("node_id", ""),
                },
            )

            runtime_error = AIRuntimeError(
                "AI Runtime 流式输出内部错误",
                code="AI_RUNTIME_STREAM_INTERNAL",
                status_code=500,
                details={
                    "reason": str(error),
                },
                retryable=False,
            )

            yield sse_encode(build_stream_error_event(runtime_error), request_id)

    return StreamingResponse(
        event_generator(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


def get_llm_service(requset:Request) -> LLMService:
    return requset.app.state.llm_service