package hooksink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Delivery is the parsed payload plus HTTP metadata made available to a
// HandlerFunc.
type Delivery struct {
	Message     *Message
	Raw         []byte
	ContentType string
	Request     *http.Request
	ReceivedAt  time.Time
}

// HandlerFunc receives a parsed webhook delivery. Returning an error lets the
// Handler write a Slack-compatible response.
type HandlerFunc func(ctx context.Context, d *Delivery) error

// Handler is an optional net/http integration layer over ParseBody.
type Handler struct {
	fn  HandlerFunc
	cfg config
}

var errNilHandlerFunc = errors.New("hooksink: nil HandlerFunc")

// New constructs a Handler. The callback is the only required argument; all
// other behavior is configured with options and has a safe default.
func New(fn HandlerFunc, opts ...Option) *Handler {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &Handler{fn: fn, cfg: cfg}
}

// ServeHTTP accepts POST requests, parses Slack-compatible payloads, invokes
// the callback, and writes Slack-compatible plain-text responses.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cfg := defaultConfig()
	var fn HandlerFunc
	if h != nil {
		cfg = h.cfg
		fn = h.fn
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		handleError(w, r, ErrMethodNotAllowed, cfg)
		return
	}

	if cfg.validator != nil {
		if err := cfg.validator(r); err != nil {
			handleError(w, r, err, cfg)
			return
		}
	}

	raw, err := readRequestBody(w, r, cfg.maxBodyBytes)
	if err != nil {
		handleError(w, r, err, cfg)
		return
	}

	msg, err := parseBody(r.Header.Get("Content-Type"), raw, cfg.formEncoded, cfg.maxAttachments)
	if err != nil {
		handleError(w, r, err, cfg)
		return
	}

	if fn == nil {
		handleError(w, r, errNilHandlerFunc, cfg)
		return
	}

	d := &Delivery{
		Message:     msg,
		Raw:         raw,
		ContentType: mediaType(r.Header.Get("Content-Type")),
		Request:     r,
		ReceivedAt:  cfg.now(),
	}
	if err := fn(r.Context(), d); err != nil {
		handleError(w, r, err, cfg)
		return
	}

	writeOK(w)
}

func readRequestBody(w http.ResponseWriter, r *http.Request, maxBytes int64) ([]byte, error) {
	var reader io.Reader = r.Body
	if maxBytes > 0 {
		reader = http.MaxBytesReader(w, r.Body, maxBytes)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, ErrPayloadTooLarge
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	return body, nil
}

func handleError(w http.ResponseWriter, r *http.Request, err error, cfg config) {
	if cfg.logger != nil {
		logHandlerError(r, err, cfg)
	}
	if cfg.errorHandler != nil {
		cfg.errorHandler(w, r, err)
		return
	}
	writeError(w, err, cfg.strictResponse)
}

func logHandlerError(r *http.Request, err error, cfg config) {
	respErr := errorResponse(err)
	if respErr.Status < http.StatusInternalServerError {
		return
	}
	if r == nil {
		cfg.logger.Error("hooksink request failed", "error", err, "status", respErr.Status, "code", respErr.Code)
		return
	}
	cfg.logger.ErrorContext(
		r.Context(),
		"hooksink request failed",
		"error", err,
		"status", respErr.Status,
		"code", respErr.Code,
	)
}
