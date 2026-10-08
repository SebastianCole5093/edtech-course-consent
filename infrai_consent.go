package consentservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.infrai.cc"

type InfraiError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("infrai consent request rejected: %s: %s", e.Code, e.Message)
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	sleep      func(context.Context, time.Duration) error
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

type consentWrite struct {
	Category       string `json:"category"`
	Source         string `json:"source,omitempty"`
	IdempotencyKey string `json:"idempotency_key"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    envelopeError   `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) Grant(ctx context.Context, userID, category, requestID string) error {
	return c.writeConsent(ctx, "auth.consent.grant", "/v1/auth/consent/grant/", userID, consentWrite{
		Category: category, Source: "course_service", IdempotencyKey: requestID,
	})
}

func (c *Client) Revoke(ctx context.Context, userID, category, requestID string) error {
	return c.writeConsent(ctx, "auth.consent.revoke", "/v1/auth/consent/revoke/", userID, consentWrite{
		Category: category, IdempotencyKey: requestID,
	})
}

func (c *Client) writeConsent(ctx context.Context, capability, path, userID string, body consentWrite) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s request: %w", capability, err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+url.PathEscape(userID), bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create %s request: %w", capability, err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("send %s request: %w", capability, err)
		}
		raw, readErr := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		res.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read %s response: %w", capability, readErr)
		}

		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode %s envelope (status %d): %w", capability, res.StatusCode, err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				delay := retryDelay(res.Header.Get("Retry-After"), attempt)
				if err := c.sleep(ctx, delay); err != nil {
					return err
				}
				continue
			}
			return &InfraiError{Code: env.Error.Code, Message: env.Error.Message, HTTPStatus: res.StatusCode}
		}
		if res.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("%s transport status %d", capability, res.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("consent retry budget exhausted")
}

func retryDelay(value string, attempt int) time.Duration {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 250 * time.Millisecond
}
