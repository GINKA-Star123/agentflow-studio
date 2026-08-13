package airuntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultTimeout = 60 * time.Second

type ClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type responseEnvelope struct {
	Data      json.RawMessage        `json:"data"`
	Error     *responseEnvelopeError `json:"error"`
	RequestID string                 `json:"request_id"`
}

type responseEnvelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func NewClient(config ClientConfig) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("ai runtime base url is required")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid ai runtime base url: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("ai runtime base url must include scheme and host")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout <= 0 {
			timeout = DefaultTimeout
		}

		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

func (c *Client) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	if c == nil {
		return nil, errors.New("ai runtime client is nil")
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	var response ChatResponse
	if err := c.postJSON(ctx, "/internal/v1/llm/chat", request, &response); err != nil {
		return nil, err
	}

	response.TokenUsage = response.TokenUsage.Normalize()

	if strings.TrimSpace(response.ResponseText()) == "" && !response.HasToolCalls() {
		return nil, errors.New("ai runtime chat response text and tool_calls are empty")
	}

	return &response, nil
}

func (c *Client) postJSON(
	ctx context.Context,
	path string,
	payload any,
	target any,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal ai runtime request failed: %w", err)
	}

	requestURL := c.baseURL + path

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL,
		bytes.NewReader(rawPayload),
	)
	if err != nil {
		return fmt.Errorf("create ai runtime request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call ai runtime failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ai runtime response failed: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"ai runtime returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return errors.New("ai runtime response body is empty")
	}

	if err := decodeJSONResponse(body, target); err != nil {
		return err
	}

	return nil
}

func decodeJSONResponse(body []byte, target any) error {
	var envelope responseEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil {
		if envelope.Error != nil {
			return fmt.Errorf(
				"ai runtime error %s: %s",
				envelope.Error.Code,
				envelope.Error.Message,
			)
		}

		if len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, target); err != nil {
				return fmt.Errorf("decode ai runtime response data failed: %w", err)
			}

			return nil
		}
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode ai runtime response failed: %w", err)
	}

	return nil
}
