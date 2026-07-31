from datetime import datetime, timezone

STARTED_AT = datetime.now(timezone.utc)

async def healthz() ->dict:
    uptime_seconds = round(
        (datetime.now(timezone.utc) - STARTED_AT).total_seconds(),
        3,
    )

    return {
        "status":"ok",
        "service":"agentflow-ai-runtime",
        "uptime_seconds":uptime_seconds
    }

async def readyz() ->dict:
    return {
        "status":"ready",
        "service":"agentflow-ai-runtime",
        "checks":{
            "runtime":"ok"
        }
    }