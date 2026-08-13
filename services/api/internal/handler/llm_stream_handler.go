package handler

import (
	"errors"
	"io"
	"net/http"

	"agentflow-studio/services/api/internal/airuntime"
	"agentflow-studio/services/api/internal/middleware"
	"agentflow-studio/services/api/internal/response"
	"agentflow-studio/services/api/internal/service"

	"github.com/gin-gonic/gin"
)

type LLMStreamHandler struct {
	llmStreamService *service.LLMStreamService
}

func NewLLMStreamHandler(
	llmStreamService *service.LLMStreamService,
) *LLMStreamHandler {
	return &LLMStreamHandler{
		llmStreamService: llmStreamService,
	}
}

func (h *LLMStreamHandler) Stream(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	var req airuntime.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			response.BindError(c, err)
			return
		}
	}

	stream, err := h.llmStreamService.StreamChat(
		c.Request.Context(),
		service.StartLLMStreamInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			Request:     req,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}
	defer func() {
		_ = stream.Close()
	}()

	prepareSSE(c)

	for {
		select {
		case <-c.Request.Context().Done():
			return

		case event, ok := <-stream.Events():
			if !ok {
				return
			}

			event.GatewayRequestID = middleware.GetRequestID(c)

			if err := writeSSE(c, event); err != nil {
				return
			}

			if event.Type == airuntime.ChatStreamEventTypeDone ||
				event.Type == airuntime.ChatStreamEventTypeError {
				return
			}

		case streamErr, ok := <-stream.Errors():
			if !ok {
				continue
			}

			event := airuntime.NewStreamErrorEvent(
				"AI_RUNTIME_STREAM_FORWARD_FAILED",
				"AI Runtime 流式转发失败",
				false,
				gin.H{
					"reason": streamErr.Error(),
				},
			)
			event.GatewayRequestID = middleware.GetRequestID(c)

			_ = writeSSE(c, event)
			return
		}
	}
}

func prepareSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func writeSSE(c *gin.Context, event airuntime.ChatStreamEvent) error {
	rawEvent, err := airuntime.EncodeSSE(event)
	if err != nil {
		return err
	}

	if _, err := c.Writer.Write(rawEvent); err != nil {
		return err
	}

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}
