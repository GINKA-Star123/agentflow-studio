from fastapi import FastAPI

from app.api.router import internal_router, root_router
from app.core.config import get_settings
from app.core.logging import configure_logging
from app.llm.factory import create_llm_service
from app.middleware.access_log import AccessLogMiddleware
from app.middleware.request_id import RequestIDMiddleware


def create_app() -> FastAPI:
    settings = get_settings()
    configure_logging(settings.app_env)

    app = FastAPI(
        title="AgentFlow AI Runtime",
        version="1.0.0",
        docs_url="/docs" if settings.app_env != "prod" else None,
        redoc_url="/redoc" if settings.app_env != "prod" else None,
    )

    app.state.settings = settings
    app.state.llm_service = create_llm_service(settings)

    app.add_middleware(AccessLogMiddleware)
    app.add_middleware(RequestIDMiddleware)

    app.include_router(root_router)
    app.include_router(internal_router)

    return app


app = create_app()