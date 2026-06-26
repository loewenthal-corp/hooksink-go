package hooksink

import (
	"log/slog"
	"net/http"
	"time"
)

// DefaultMaxBodyBytes is the default request body limit used by Parse and
// Handler.
const DefaultMaxBodyBytes int64 = 1 << 20

// DefaultMaxAttachments is the default attachment count limit, matching Slack's
// documented incoming webhook limit.
const DefaultMaxAttachments = 100

// Option configures a Handler.
type Option func(*config)

type config struct {
	maxBodyBytes   int64
	formEncoded    bool
	strictResponse bool
	errorHandler   func(http.ResponseWriter, *http.Request, error)
	validator      func(*http.Request) error
	maxAttachments int
	logger         *slog.Logger
	now            func() time.Time
}

func defaultConfig() config {
	return config{
		maxBodyBytes:   DefaultMaxBodyBytes,
		formEncoded:    true,
		strictResponse: true,
		maxAttachments: DefaultMaxAttachments,
		now:            time.Now,
	}
}

// WithMaxBodyBytes sets the maximum request body size. Values less than or
// equal to zero disable the size limit.
func WithMaxBodyBytes(n int64) Option {
	return func(c *config) {
		c.maxBodyBytes = n
	}
}

// WithFormEncoded enables or disables legacy
// application/x-www-form-urlencoded payload=<json> parsing.
func WithFormEncoded(enabled bool) Option {
	return func(c *config) {
		c.formEncoded = enabled
	}
}

// WithStrictResponses enables or disables Slack-compatible plain-text response
// codes. It defaults to true.
func WithStrictResponses(enabled bool) Option {
	return func(c *config) {
		c.strictResponse = enabled
	}
}

// WithErrorHandler overrides Handler's default failure responses.
func WithErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) Option {
	return func(c *config) {
		c.errorHandler = fn
	}
}

// WithValidator installs a pre-parse request validator. It can enforce
// per-path tokens, shared secrets, or any application-specific authorization.
func WithValidator(fn func(*http.Request) error) Option {
	return func(c *config) {
		c.validator = fn
	}
}

// WithMaxAttachments sets the maximum number of attachments accepted in one
// payload. A value of zero disables the attachment limit.
func WithMaxAttachments(n int) Option {
	return func(c *config) {
		if n < 0 {
			n = 0
		}
		c.maxAttachments = n
	}
}

// WithLogger sets a logger for unexpected handler failures. Nil disables
// logging.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

// WithNow sets the clock used to populate Delivery.ReceivedAt. Nil leaves the
// default clock unchanged.
func WithNow(fn func() time.Time) Option {
	return func(c *config) {
		if fn != nil {
			c.now = fn
		}
	}
}
