package hooksink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

// ParseBody parses a Slack-compatible incoming webhook body without depending
// on net/http.
func ParseBody(contentType string, body []byte) (*Message, error) {
	cfg := defaultConfig()
	return parseBody(contentType, body, cfg.formEncoded, cfg.maxAttachments)
}

// Parse reads and parses a Slack-compatible incoming webhook request. It
// accepts only POST requests and consumes r.Body.
func Parse(r *http.Request) (*Message, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: nil request", ErrInvalidPayload)
	}
	if r.Method != http.MethodPost {
		return nil, ErrMethodNotAllowed
	}

	cfg := defaultConfig()
	body, err := readAllLimited(r.Body, cfg.maxBodyBytes)
	if err != nil {
		return nil, err
	}

	return parseBody(r.Header.Get("Content-Type"), body, cfg.formEncoded, cfg.maxAttachments)
}

func parseBody(contentType string, body []byte, formEncoded bool, maxAttachments int) (*Message, error) {
	raw := body
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, ErrInvalidPayload
	}

	if mediaType(contentType) == "application/x-www-form-urlencoded" {
		if !formEncoded {
			return nil, ErrInvalidPayload
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
		}
		payload := values.Get("payload")
		if strings.TrimSpace(payload) == "" {
			return nil, ErrInvalidPayload
		}
		raw = []byte(payload)
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, ErrInvalidPayload
	}

	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	if err := validateMessage(&msg, maxAttachments); err != nil {
		return nil, err
	}

	return &msg, nil
}

func validateMessage(msg *Message, maxAttachments int) error {
	if maxAttachments > 0 && len(msg.Attachments) > maxAttachments {
		return ErrTooManyAttachments
	}

	hasText := strings.TrimSpace(msg.Text) != ""
	hasBlocks := len(msg.Blocks.BlockSet) > 0
	hasAttachments := len(msg.Attachments) > 0
	if !hasText && !hasBlocks && !hasAttachments {
		return ErrNoText
	}

	return nil
}

func readAllLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if r == nil {
		return nil, ErrInvalidPayload
	}
	if maxBytes <= 0 || maxBytes == math.MaxInt64 {
		body, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
		}
		return body, nil
	}

	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, ErrPayloadTooLarge
	}
	return body, nil
}

func mediaType(contentType string) string {
	parsed, _, err := mime.ParseMediaType(contentType)
	if err == nil {
		return strings.ToLower(parsed)
	}

	beforeParams, _, _ := strings.Cut(contentType, ";")
	return strings.ToLower(strings.TrimSpace(beforeParams))
}
