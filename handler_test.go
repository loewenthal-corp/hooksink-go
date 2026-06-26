package hooksink_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func TestHandler_ResponseContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		body       string
		contentTyp string
		opts       []hooksink.Option
		fn         hooksink.HandlerFunc
		wantStatus int
		wantBody   string
		wantAllow  string
	}{
		{
			name:       "success",
			method:     http.MethodPost,
			body:       `{"text":"hello"}`,
			contentTyp: "application/json",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "malformed json",
			method:     http.MethodPost,
			body:       `{"text":`,
			contentTyp: "application/json",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid_payload",
		},
		{
			name:       "empty body",
			method:     http.MethodPost,
			body:       "",
			contentTyp: "application/json",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid_payload",
		},
		{
			name:       "no text",
			method:     http.MethodPost,
			body:       `{"username":"bot"}`,
			contentTyp: "application/json",
			wantStatus: http.StatusBadRequest,
			wantBody:   "no_text",
		},
		{
			name:       "too many attachments with configured max",
			method:     http.MethodPost,
			body:       attachmentPayload(2),
			contentTyp: "application/json",
			opts:       []hooksink.Option{hooksink.WithMaxAttachments(1)},
			wantStatus: http.StatusBadRequest,
			wantBody:   "too_many_attachments",
		},
		{
			name:       "body too large",
			method:     http.MethodPost,
			body:       `{"text":"this is longer than the configured maximum"}`,
			contentTyp: "application/json",
			opts:       []hooksink.Option{hooksink.WithMaxBodyBytes(8)},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   "payload_too_large",
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			body:       `{"text":"hello"}`,
			contentTyp: "application/json",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method_not_allowed",
			wantAllow:  http.MethodPost,
		},
		{
			name:       "callback response error",
			method:     http.MethodPost,
			body:       `{"text":"hello"}`,
			contentTyp: "application/json",
			fn: func(context.Context, *hooksink.Delivery) error {
				return hooksink.NewResponseError(http.StatusTeapot, "teapot")
			},
			wantStatus: http.StatusTeapot,
			wantBody:   "teapot",
		},
		{
			name:       "callback generic error",
			method:     http.MethodPost,
			body:       `{"text":"hello"}`,
			contentTyp: "application/json",
			fn: func(context.Context, *hooksink.Delivery) error {
				return errors.New("database unavailable")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "rollup_error",
		},
		{
			name:       "form disabled",
			method:     http.MethodPost,
			body:       "payload=%7B%22text%22%3A%22hello%22%7D",
			contentTyp: "application/x-www-form-urlencoded",
			opts:       []hooksink.Option{hooksink.WithFormEncoded(false)},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid_payload",
		},
		{
			name:       "non strict parse response",
			method:     http.MethodPost,
			body:       `{"text":`,
			contentTyp: "application/json",
			opts:       []hooksink.Option{hooksink.WithStrictResponses(false)},
			wantStatus: http.StatusBadRequest,
			wantBody:   "bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fn := tt.fn
			if fn == nil {
				fn = func(context.Context, *hooksink.Delivery) error { return nil }
			}
			h := hooksink.New(fn, tt.opts...)
			req := httptest.NewRequest(tt.method, "/services/T/B/X", strings.NewReader(tt.body))
			if tt.contentTyp != "" {
				req.Header.Set("Content-Type", tt.contentTyp)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body %q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Fatalf("Content-Type = %q, want text/plain; charset=utf-8", got)
			}
			if tt.wantAllow != "" && rec.Header().Get("Allow") != tt.wantAllow {
				t.Fatalf("Allow = %q, want %q", rec.Header().Get("Allow"), tt.wantAllow)
			}
		})
	}
}

func TestHandler_DeliveryMetadata(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)
	var seen bool
	h := hooksink.New(
		func(ctx context.Context, d *hooksink.Delivery) error {
			seen = true
			if ctx != d.Request.Context() {
				t.Fatal("callback context is not request context")
			}
			if d.Message.Text != "hello" {
				t.Fatalf("Message.Text = %q, want hello", d.Message.Text)
			}
			if string(d.Raw) != `{"text":"hello"}` {
				t.Fatalf("Raw = %q, want original body", string(d.Raw))
			}
			if d.ContentType != "application/json" {
				t.Fatalf("ContentType = %q, want application/json", d.ContentType)
			}
			if d.Request.URL.Path != "/services/T/B/X" {
				t.Fatalf("path = %q, want /services/T/B/X", d.Request.URL.Path)
			}
			if !d.ReceivedAt.Equal(fixedNow) {
				t.Fatalf("ReceivedAt = %s, want %s", d.ReceivedAt, fixedNow)
			}
			return nil
		},
		hooksink.WithNow(func() time.Time { return fixedNow }),
	)

	req := httptest.NewRequest(http.MethodPost, "/services/T/B/X", strings.NewReader(`{"text":"hello"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !seen {
		t.Fatal("callback was not called")
	}
}

func TestHandler_ValidatorShortCircuitsBeforeParse(t *testing.T) {
	t.Parallel()

	var called atomic.Bool
	h := hooksink.New(
		func(context.Context, *hooksink.Delivery) error {
			called.Store(true)
			return nil
		},
		hooksink.WithValidator(func(*http.Request) error {
			return hooksink.ErrInvalidToken
		}),
	)
	req := httptest.NewRequest(http.MethodPost, "/services/T/B/wrong", strings.NewReader(`{"text":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if rec.Body.String() != "invalid_token" {
		t.Fatalf("body = %q, want invalid_token", rec.Body.String())
	}
	if called.Load() {
		t.Fatal("callback was called")
	}
}

func TestHandler_WithErrorHandler(t *testing.T) {
	t.Parallel()

	h := hooksink.New(
		func(context.Context, *hooksink.Delivery) error { return nil },
		hooksink.WithErrorHandler(func(w http.ResponseWriter, _ *http.Request, err error) {
			if !errors.Is(err, hooksink.ErrInvalidPayload) {
				t.Fatalf("error = %v, want ErrInvalidPayload", err)
			}
			w.WriteHeader(499)
			_, _ = w.Write([]byte("custom"))
		}),
	)
	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"text":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != 499 {
		t.Fatalf("status = %d, want 499", rec.Code)
	}
	if rec.Body.String() != "custom" {
		t.Fatalf("body = %q, want custom", rec.Body.String())
	}
}

func TestHandler_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	var count atomic.Int64
	h := hooksink.New(func(context.Context, *hooksink.Delivery) error {
		count.Add(1)
		return nil
	})

	const requests = 64
	var wg sync.WaitGroup
	wg.Add(requests)
	for range requests {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"text":"hello"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
		}()
	}
	wg.Wait()

	if count.Load() != requests {
		t.Fatalf("callback count = %d, want %d", count.Load(), requests)
	}
}
