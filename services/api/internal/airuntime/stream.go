package airuntime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxSSELineSize = 1024 * 1024

type ChatStream struct {
	events <-chan ChatStreamEvent
	errs   <-chan error
	close  func() error
}

func (s *ChatStream) Events() <-chan ChatStreamEvent {
	return s.events
}

func (s *ChatStream) Errors() <-chan error {
	return s.errs
}

func (s *ChatStream) Close() error {
	if s == nil || s.close == nil {
		return nil
	}

	return s.close()
}

func (c *Client) StreamChat(ctx context.Context, request ChatRequest) (*ChatStream, error) {
	if c == nil {
		return nil, errors.New("ai runtime client is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	rawPayload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal ai runtime stream request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/internal/v1/llm/stream",
		bytes.NewReader(rawPayload),
	)
	if err != nil {
		return nil, fmt.Errorf("create ai runtime stream request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ai runtime stream failed: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer func() {
			_ = resp.Body.Close()
		}()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("read ai runtime stream error response failed: %w", readErr)
		}

		return nil, fmt.Errorf(
			"ai runtime stream returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	events := make(chan ChatStreamEvent)
	errs := make(chan error, 1)

	go decodeSSEStream(ctx, resp.Body, events, errs)

	return &ChatStream{
		events: events,
		errs:   errs,
		close:  resp.Body.Close,
	}, nil
}

func decodeSSEStream(
	ctx context.Context,
	body io.ReadCloser,
	events chan<- ChatStreamEvent,
	errs chan<- error,
) {
	defer close(events)
	defer close(errs)
	defer func() {
		_ = body.Close()
	}()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), maxSSELineSize)

	eventName := ""
	dataLines := []string{}

	flush := func() bool {
		if len(dataLines) == 0 {
			eventName = ""
			return true
		}

		rawData := strings.Join(dataLines, "\n")
		dataLines = nil

		event, err := decodeSSEEvent(eventName, rawData)
		eventName = ""

		if err != nil {
			sendStreamError(ctx, errs, err)
			return false
		}

		select {
		case <-ctx.Done():
			return false
		case events <- event:
			return true
		}
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		if line == "" {
			if ok := flush(); !ok {
				return
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		value = strings.TrimPrefix(value, " ")

		switch field {
		case "event":
			eventName = strings.TrimSpace(value)

		case "data":
			dataLines = append(dataLines, value)
		}
	}

	if len(dataLines) > 0 {
		_ = flush()
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		sendStreamError(
			ctx,
			errs,
			fmt.Errorf("read ai runtime sse stream failed: %w", err),
		)
	}
}

func decodeSSEEvent(eventName string, rawData string) (ChatStreamEvent, error) {
	rawData = strings.TrimSpace(rawData)
	if rawData == "" {
		return ChatStreamEvent{}, errors.New("sse data is empty")
	}

	if rawData == "[DONE]" {
		return ChatStreamEvent{
			Type:         ChatStreamEventTypeDone,
			FinishReason: "stop",
		}, nil
	}

	var event ChatStreamEvent
	if err := json.Unmarshal([]byte(rawData), &event); err != nil {
		return ChatStreamEvent{}, fmt.Errorf("decode ai runtime sse data failed: %w", err)
	}

	if event.Type == "" && eventName != "" {
		event.Type = ChatStreamEventType(eventName)
	}

	if event.Type == "" {
		event.Type = ChatStreamEventType("message")
	}

	return event, nil
}

func sendStreamError(ctx context.Context, errs chan<- error, err error) {
	select {
	case <-ctx.Done():
	case errs <- err:
	default:
	}
}
