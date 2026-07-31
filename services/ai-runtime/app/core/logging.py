import json
import logging
import sys
from datetime import datetime, timezone
from typing import Any

class JsonFormatter(logging.Formatter):
    extra_fields = (
        "request_id",
        "method",
        "path",
        "status_code",
        "latency_ms",
        "client_ip",
        "trace_id",
        "run_id",
        "node_id"
    )

    def format(self,record:logging.LogRecord) ->str:
        payload: dict[str,Any] = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "level":record.levelname.lower(),
            "logger":record.name,
            "message":record.getMessage(),
        }

        for field in self.extra_fields:
            value = getattr(record, field, None)
            if value is not None:
                payload[field] = value

        if record.exc_info:
            payload["exception"] = self.formatException(record.exc_info)

        return json.dumps(payload,ensure_ascii=False)

def configure_logging(app_env:str) ->None:
    root_logger = logging.getLogger()
    root_logger.handlers.clear()

    if app_env == "prod":
        root_logger.setLevel(logging.INFO)

    else:
        root_logger.setLevel(logging.DEBUG)

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JsonFormatter())

    root_logger.addHandler(handler)

    logging.getLogger("uvicorn.access").disabled = True