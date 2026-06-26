package hooksink_test

import (
	"encoding/json"
	"testing"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func TestMessage_UnmarshalJSONPreservesExtra(t *testing.T) {
	t.Parallel()

	var msg hooksink.Message
	err := json.Unmarshal([]byte(`{
		"text": "hello",
		"future_field": {"nested": true},
		"retry_count": 3
	}`), &msg)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if msg.Text != "hello" {
		t.Fatalf("Text = %q, want hello", msg.Text)
	}
	if len(msg.Extra) != 2 {
		t.Fatalf("len(Extra) = %d, want 2", len(msg.Extra))
	}

	var retryCount int
	if err := json.Unmarshal(msg.Extra["retry_count"], &retryCount); err != nil {
		t.Fatalf("unmarshal retry_count: %v", err)
	}
	if retryCount != 3 {
		t.Fatalf("retry_count = %d, want 3", retryCount)
	}

	var future struct {
		Nested bool `json:"nested"`
	}
	if err := json.Unmarshal(msg.Extra["future_field"], &future); err != nil {
		t.Fatalf("unmarshal future_field: %v", err)
	}
	if !future.Nested {
		t.Fatal("future_field.nested = false, want true")
	}
}

func TestMessage_MrkdwnPointerSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		wantNil    bool
		wantMrkdwn bool
	}{
		{
			name:       "absent",
			body:       `{"text":"hello"}`,
			wantNil:    true,
			wantMrkdwn: false,
		},
		{
			name:       "false",
			body:       `{"text":"hello","mrkdwn":false}`,
			wantNil:    false,
			wantMrkdwn: false,
		},
		{
			name:       "true",
			body:       `{"text":"hello","mrkdwn":true}`,
			wantNil:    false,
			wantMrkdwn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, err := hooksink.ParseBody("application/json", []byte(tt.body))
			if err != nil {
				t.Fatalf("ParseBody() error = %v", err)
			}
			if gotNil := msg.Mrkdwn == nil; gotNil != tt.wantNil {
				t.Fatalf("Mrkdwn nil = %v, want %v", gotNil, tt.wantNil)
			}
			if msg.Mrkdwn != nil && *msg.Mrkdwn != tt.wantMrkdwn {
				t.Fatalf("*Mrkdwn = %v, want %v", *msg.Mrkdwn, tt.wantMrkdwn)
			}
		})
	}
}

func TestMessage_MarshalJSONIncludesExtra(t *testing.T) {
	t.Parallel()

	mrkdwn := false
	msg := hooksink.Message{
		Text:   "hello",
		Mrkdwn: &mrkdwn,
		Extra: map[string]json.RawMessage{
			"future_field": json.RawMessage(`{"nested":true}`),
			"text":         json.RawMessage(`"ignored"`),
		},
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal marshaled message: %v", err)
	}
	if string(got["text"]) != `"hello"` {
		t.Fatalf("marshaled text = %s, want %q", got["text"], `"hello"`)
	}
	if string(got["mrkdwn"]) != `false` {
		t.Fatalf("marshaled mrkdwn = %s, want false", got["mrkdwn"])
	}
	if string(got["future_field"]) != `{"nested":true}` {
		t.Fatalf("marshaled future_field = %s, want nested object", got["future_field"])
	}
}
