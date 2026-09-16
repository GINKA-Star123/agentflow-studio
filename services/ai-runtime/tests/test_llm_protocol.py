from __future__ import annotations

import json
import sys
import unittest
from pathlib import Path
from types import SimpleNamespace

PROJECT_ROOT = Path(__file__).resolve().parents[1]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from app.core.config import Settings
from app.llm.model_params import normalize_model_parameters
from app.llm.providers.mock import MockLLMProvider
from app.llm.providers.openai_compatible import (
    apply_tool_call_delta,
    build_chat_response,
    build_openai_messages,
    build_openai_tool_options,
    build_tool_calls_from_stream_state,
    parse_openai_tool_call_delta,
)
from app.llm.schemas import ChatMessage, ChatRequest, ChatRole, ToolCall, TokenUsage
from app.llm.service import LLMService, normalize_provider_name
from app.tools.schemas import ToolCallResponse, ToolCallStatus, build_tool_result_message
from pydantic import ValidationError


class ChatSchemaTests(unittest.TestCase):
    def test_tool_message_requires_tool_call_id(self) -> None:
        with self.assertRaises(ValidationError):
            ChatMessage(role=ChatRole.tool, content="tool result")

    def test_build_tool_result_message_serializes_payload(self) -> None:
        response = ToolCallResponse(
            tool_call_id="call_1",
            tool_name="search_docs",
            status=ToolCallStatus.succeeded,
            result={"items": [{"id": 1}]},
            metadata={"source": "mock"},
            latency_ms=12,
        )

        message = build_tool_result_message(response)
        payload = json.loads(message.content)

        self.assertEqual(message.role, "tool")
        self.assertEqual(message.tool_call_id, "call_1")
        self.assertEqual(payload["status"], "succeeded")
        self.assertEqual(payload["result"]["items"][0]["id"], 1)
        self.assertEqual(payload["metadata"]["source"], "mock")


class ModelParameterTests(unittest.TestCase):
    def test_normalize_model_parameters_clamps_max_tokens(self) -> None:
        settings = Settings(
            llm_default_temperature=0.6,
            llm_default_max_tokens=2048,
            llm_max_tokens_limit=1024,
        )
        request = ChatRequest(
            provider="mock",
            model="demo",
            messages=[
                ChatMessage(role=ChatRole.user, content="hello"),
            ],
            temperature=None,
            max_tokens=2048,
        )

        params = normalize_model_parameters(request, settings)

        self.assertEqual(params.temperature, 0.6)
        self.assertEqual(params.max_tokens, 1024)
        self.assertEqual(params.to_openai_options()["max_tokens"], 1024)


class OpenAICompatibleProtocolTests(unittest.TestCase):
    def test_build_openai_messages_preserves_tool_fields(self) -> None:
        request = ChatRequest(
            provider="openai-compatible",
            model="demo",
            messages=[
                ChatMessage(role=ChatRole.system, content="system"),
                ChatMessage(role=ChatRole.user, content="user"),
                ChatMessage(
                    role=ChatRole.assistant,
                    content="",
                    tool_calls=[
                        ToolCall(
                            id="call_1",
                            type="function",
                            function={
                                "name": "search_docs",
                                "arguments": '{"query":"phase5"}',
                            },
                        ),
                    ],
                ),
                ChatMessage(
                    role=ChatRole.tool,
                    content='{"items":[1,2,3]}',
                    tool_call_id="call_1",
                ),
            ],
            tools=[
                {
                    "type": "function",
                    "function": {
                        "name": "search_docs",
                        "description": "Search docs",
                    },
                },
            ],
            tool_choice="auto",
        )

        messages = build_openai_messages(request)
        options = build_openai_tool_options(request)

        self.assertEqual(messages[0]["role"], "system")
        self.assertEqual(messages[2]["tool_calls"][0]["function"]["name"], "search_docs")
        self.assertEqual(messages[3]["tool_call_id"], "call_1")
        self.assertEqual(options["tool_choice"], "auto")
        self.assertEqual(options["tools"][0]["function"]["name"], "search_docs")

    def test_build_chat_response_extracts_tool_calls_and_usage(self) -> None:
        request = ChatRequest(
            provider="openai-compatible",
            model="demo",
            messages=[
                ChatMessage(role=ChatRole.user, content="hello"),
            ],
        )
        parameters = normalize_model_parameters(request, Settings())
        response = SimpleNamespace(
            id="resp_1",
            choices=[
                SimpleNamespace(
                    finish_reason="stop",
                    message=SimpleNamespace(
                        content="hello world",
                        tool_calls=[
                            SimpleNamespace(
                                id="call_1",
                                function=SimpleNamespace(
                                    name="search_docs",
                                    arguments='{"query":"phase5"}',
                                ),
                            )
                        ],
                    ),
                )
            ],
            usage=SimpleNamespace(
                prompt_tokens=10,
                completion_tokens=4,
                total_tokens=0,
            ),
        )

        chat_response = build_chat_response(
            response=response,
            request=request,
            provider_name="openai-compatible",
            base_url="http://localhost:1234/v1",
            parameters=parameters,
            attempt=1,
            latency_ms=23,
        )

        self.assertEqual(chat_response.text, "hello world")
        self.assertEqual(chat_response.tool_calls[0].id, "call_1")
        self.assertEqual(chat_response.token_usage.total_tokens, 14)
        self.assertEqual(chat_response.raw["latency_ms"], 23)

    def test_stream_tool_call_state_is_assembled(self) -> None:
        state: dict[int, dict[str, object]] = {}

        first_delta = parse_openai_tool_call_delta(
            SimpleNamespace(
                index=0,
                id="call_1",
                type="function",
                function=SimpleNamespace(
                    name="search_docs",
                    arguments='{"query":"phase',
                ),
            )
        )
        second_delta = parse_openai_tool_call_delta(
            SimpleNamespace(
                index=0,
                id="",
                type="function",
                function=SimpleNamespace(
                    name="",
                    arguments='5"}',
                ),
            )
        )

        self.assertIsNotNone(first_delta)
        self.assertIsNotNone(second_delta)

        apply_tool_call_delta(state, first_delta)  # type: ignore[arg-type]
        apply_tool_call_delta(state, second_delta)  # type: ignore[arg-type]

        tool_calls = build_tool_calls_from_stream_state(state)

        self.assertEqual(len(tool_calls), 1)
        self.assertEqual(tool_calls[0].id, "call_1")
        self.assertEqual(tool_calls[0].function.name, "search_docs")
        self.assertEqual(tool_calls[0].function.arguments, '{"query":"phase5"}')


class ProviderRegistryTests(unittest.TestCase):
    def test_llm_service_registers_and_resolves_mock_provider(self) -> None:
        service = LLMService()
        service.register_provider(MockLLMProvider())

        self.assertEqual(normalize_provider_name(" MOCK "), "mock")
        self.assertEqual(service.list_provider_names(), ["mock"])
        self.assertEqual(service.get_provider(" MOCK ").name, "mock")


class MockProviderTests(unittest.IsolatedAsyncioTestCase):
    async def test_mock_provider_chat_and_stream(self) -> None:
        provider = MockLLMProvider()
        request = ChatRequest(
            provider="mock",
            model="mock-chat",
            messages=[
                ChatMessage(role=ChatRole.user, content="hello phase 5"),
            ],
        )

        response = await provider.chat(request)
        self.assertEqual(response.message.role, ChatRole.assistant)
        self.assertIn("mock-chat", response.text)
        self.assertGreater(response.token_usage.total_tokens, 0)

        events = [event async for event in provider.stream_chat(request)]
        event_types = [event.type.value for event in events]

        self.assertIn("start", event_types)
        self.assertIn("delta", event_types)
        self.assertIn("usage", event_types)
        self.assertEqual(event_types[-1], "done")
        self.assertEqual(events[-1].finish_reason, "stop")
        self.assertIsNotNone(events[-1].token_usage)


if __name__ == "__main__":
    unittest.main()
