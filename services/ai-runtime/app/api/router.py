from fastapi import APIRouter

from app.api import health, llm, tools

root_router = APIRouter()
root_router.add_api_route(
    "/healthz",
    health.healthz,
    methods=["GET"],
    tags=["health"],
)

root_router.add_api_route(
    "/readyz",
    health.readyz,
    methods=["GET"],
    tags=["health"],
)

internal_router = APIRouter(prefix="/internal/v1")
internal_router.add_api_route(
    "/health",
    health.healthz,
    methods=["GET"],
    tags=["health"],
)
internal_router.include_router(llm.router)
internal_router.include_router(tools.router)