import logging
import time

from starlette.middleware.base import BaseHTTPMiddleware,RequestResponseEndpoint
from starlette.requests import Request
from starlette.responses import Response

class AccessLogMiddleware(BaseHTTPMiddleware):
    def __init__(self,app) ->None:
        super().__init__(app)
        self.logger = logging.getLogger("ai_runtime.access")

    async def dispatch(
            self,
            request:Request,
            call_next:RequestResponseEndpoint,
    ) ->Response:
        start_time = time.perf_counter()
        request_id = getattr(request.state, "request_id", "")

        try:
            response = await call_next(request)

        except Exception:
            latency_ms = round((time.perf_counter() - start_time) * 1000, 2)

            self.logger.error(
                "http 请求异常",
                extra = {
                    "request_id":request_id,
                    "method":request.method,
                    "path":request.url.path,
                    "status_code":500,
                    "latency_ms":latency_ms,
                    "client_ip":request.client.host if request.client else "",
                },
            )
            raise

        latency_ms = round((time.perf_counter() - start_time) * 1000, 2)

        self.logger.info(
            "http 请求完成",
            extra={
                "request_id": request_id,
                "method": request.method,
                "path": request.url.path,
                "status_code": response.status_code,
                "latency_ms": latency_ms,
                "client_ip": request.client.host if request.client else "",
            },
        )

        return response