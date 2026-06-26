package hooksink_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	hooksink "github.com/loewenthal-corp/hooksink-go"
	"github.com/slack-go/slack"
)

const validRichPayload = `{
	"text": "deploy finished",
	"blocks": [
		{
			"type": "section",
			"text": {"type": "mrkdwn", "text": "*deploy finished*"}
		}
	],
	"attachments": [
		{"fallback": "fallback text", "text": "attachment text", "color": "good"}
	],
	"username": "ci",
	"channel": "#deploys"
}`

func TestParseBody(t *testing.T) {
	t.Parallel()

	form := url.Values{"payload": {validRichPayload}}.Encode()

	tests := []struct {
		name        string
		contentType string
		body        string
		wantErr     error
		assert      func(*testing.T, *hooksink.Message)
	}{
		{
			name:        "application json with blocks and attachments",
			contentType: "application/json",
			body:        validRichPayload,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if msg.Text != "deploy finished" {
					t.Fatalf("Text = %q, want deploy finished", msg.Text)
				}
				if len(msg.Blocks.BlockSet) != 1 {
					t.Fatalf("blocks = %d, want 1", len(msg.Blocks.BlockSet))
				}
				if msg.Blocks.BlockSet[0].BlockType() != slack.MBTSection {
					t.Fatalf("block type = %s, want section", msg.Blocks.BlockSet[0].BlockType())
				}
				if len(msg.Attachments) != 1 {
					t.Fatalf("attachments = %d, want 1", len(msg.Attachments))
				}
				if msg.Attachments[0].Text != "attachment text" {
					t.Fatalf("attachment text = %q, want attachment text", msg.Attachments[0].Text)
				}
			},
		},
		{
			name:        "form encoded payload",
			contentType: "application/x-www-form-urlencoded",
			body:        form,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if msg.Text != "deploy finished" {
					t.Fatalf("Text = %q, want deploy finished", msg.Text)
				}
				if len(msg.Blocks.BlockSet) != 1 || len(msg.Attachments) != 1 {
					t.Fatalf("blocks = %d attachments = %d, want 1 and 1", len(msg.Blocks.BlockSet), len(msg.Attachments))
				}
			},
		},
		{
			name:        "json with charset",
			contentType: "application/json; charset=utf-8",
			body:        `{"text":"hello"}`,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if msg.Text != "hello" {
					t.Fatalf("Text = %q, want hello", msg.Text)
				}
			},
		},
		{
			name:        "missing content type attempts json",
			contentType: "",
			body:        `{"text":"hello"}`,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if msg.Text != "hello" {
					t.Fatalf("Text = %q, want hello", msg.Text)
				}
			},
		},
		{
			name:        "only blocks is valid",
			contentType: "application/json",
			body:        `{"blocks":[{"type":"divider"}]}`,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if len(msg.Blocks.BlockSet) != 1 {
					t.Fatalf("blocks = %d, want 1", len(msg.Blocks.BlockSet))
				}
			},
		},
		{
			name:        "only attachments is valid",
			contentType: "application/json",
			body:        `{"attachments":[{"text":"attachment only"}]}`,
			assert: func(t *testing.T, msg *hooksink.Message) {
				t.Helper()
				if len(msg.Attachments) != 1 {
					t.Fatalf("attachments = %d, want 1", len(msg.Attachments))
				}
			},
		},
		{
			name:        "empty body",
			contentType: "application/json",
			body:        "",
			wantErr:     hooksink.ErrInvalidPayload,
		},
		{
			name:        "whitespace body",
			contentType: "application/json",
			body:        " \n\t ",
			wantErr:     hooksink.ErrInvalidPayload,
		},
		{
			name:        "malformed json",
			contentType: "application/json",
			body:        `{"text":`,
			wantErr:     hooksink.ErrInvalidPayload,
		},
		{
			name:        "none of text blocks attachments",
			contentType: "application/json",
			body:        `{"username":"bot"}`,
			wantErr:     hooksink.ErrNoText,
		},
		{
			name:        "too many attachments",
			contentType: "application/json",
			body:        attachmentPayload(101),
			wantErr:     hooksink.ErrTooManyAttachments,
		},
		{
			name:        "form payload missing",
			contentType: "application/x-www-form-urlencoded",
			body:        "not_payload=%7B%7D",
			wantErr:     hooksink.ErrInvalidPayload,
		},
		{
			name:        "json null is invalid payload",
			contentType: "application/json",
			body:        `null`,
			wantErr:     hooksink.ErrInvalidPayload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, err := hooksink.ParseBody(tt.contentType, []byte(tt.body))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseBody() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseBody() error = %v", err)
			}
			if msg == nil {
				t.Fatal("ParseBody() message is nil")
			}
			if tt.assert != nil {
				tt.assert(t, msg)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		method  string
		body    string
		wantErr error
	}{
		{
			name:   "post parses",
			method: http.MethodPost,
			body:   `{"text":"hello"}`,
		},
		{
			name:    "get rejected",
			method:  http.MethodGet,
			body:    `{"text":"hello"}`,
			wantErr: hooksink.ErrMethodNotAllowed,
		},
		{
			name:    "default body limit",
			method:  http.MethodPost,
			body:    strings.Repeat("x", int(hooksink.DefaultMaxBodyBytes)+1),
			wantErr: hooksink.ErrPayloadTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/hook", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			msg, err := hooksink.Parse(req)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Parse() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if msg.Text != "hello" {
				t.Fatalf("Text = %q, want hello", msg.Text)
			}
		})
	}
}

func attachmentPayload(n int) string {
	var b strings.Builder
	b.WriteString(`{"attachments":[`)
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"text":"attachment"}`)
	}
	b.WriteString(`]}`)
	return b.String()
}
