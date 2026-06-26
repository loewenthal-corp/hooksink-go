# hooksink-go

Receive Slack-compatible incoming webhooks in Go.

`hooksink-go` is the server side of Slack's incoming webhook format. It lets you
mint URLs that third-party tools already know how to call, accepts the same JSON
or legacy `payload=<json>` form body they would send to `hooks.slack.com`, and
turns the payload into typed Go values.

It is not a Slack client, does not post messages to Slack, and does not mock the
Slack Web API.

## Install

```sh
go get github.com/loewenthal-corp/hooksink-go
```

The core package depends only on the Go standard library plus
`github.com/slack-go/slack` for Block Kit and attachment types.

## Development

The repo uses Hermit and Taskfile for a pinned, repeatable dev loop:

```sh
source bin/activate-hermit
task do
```

`task do` formats with gofumpt, tidies all modules, builds the root package and
router examples, runs golangci-lint/actionlint/typos, runs race+coverage tests,
checks Go 1.22 compatibility, and runs govulncheck/gitleaks/zizmor. Use
`task --list` to see focused tasks such as `task test::fuzz` and `task security`.

## Parse Only

Use `Parse` when you own the HTTP handler and want to decide how errors map to
responses.

```go
package main

import (
	"errors"
	"net/http"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	msg, err := hooksink.Parse(r)
	if err != nil {
		var responseErr *hooksink.ResponseError
		if errors.As(err, &responseErr) {
			http.Error(w, responseErr.Code, responseErr.Status)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = msg // route, enqueue, store, fan out, or inspect the payload here.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
```

For non-HTTP code, call the pure parser directly:

```go
msg, err := hooksink.ParseBody("application/json", []byte(`{"text":"hello"}`))
```

## Handler

Use `New` when you want Slack-compatible responses handled for you. A nil
callback error writes `200 ok`; a returned `*ResponseError` writes that exact
status and plain-text body; any other callback error writes `500 rollup_error`.

```go
h := hooksink.New(func(ctx context.Context, d *hooksink.Delivery) error {
	log.Printf(
		"from %s: %q (%d blocks)",
		d.Request.URL.Path,
		d.Message.Text,
		len(d.Message.Blocks.BlockSet),
	)
	return nil
})
```

`Delivery` includes the parsed message, original body bytes, normalized content
type, the request, and the receive time.

Useful options:

```go
h := hooksink.New(
	fn,
	hooksink.WithMaxBodyBytes(1<<20),
	hooksink.WithFormEncoded(true),
	hooksink.WithMaxAttachments(100),
	hooksink.WithValidator(func(r *http.Request) error {
		// Validate path tokens, shared secrets, or tenant routing here.
		return nil
	}),
)
```

## Router Matrix

`Handler` implements `http.Handler`, so router integration stays boring.

```go
// stdlib, Go 1.22+ pattern routing
mux := http.NewServeMux()
mux.Handle("POST /services/{team}/{bot}/{token}", h)

// chi
r.Method(http.MethodPost, "/hook/{id}", h)

// gorilla/mux
m.Handle("/hook/{id}", h).Methods(http.MethodPost)

// gin
g.POST("/hook/:id", gin.WrapH(h))

// echo
e.POST("/hook/:id", echo.WrapHandler(h))
```

Runnable examples are in:

- `examples/stdlib`
- `examples/chi`
- `examples/gorilla`
- `examples/gin`
- `examples/echo`

## Wire Format

Accepted payload shapes:

- `application/json`: raw Slack-compatible JSON.
- `application/x-www-form-urlencoded`: legacy `payload=<url-encoded JSON>`.
- Missing or unexpected content type: parsed leniently as JSON.

Charset parameters such as `application/json; charset=utf-8` are handled with
`mime.ParseMediaType`.

Strict handler responses are plain text:

| Condition | Status | Body |
| --- | ---: | --- |
| Success | `200` | `ok` |
| Empty, whitespace-only, malformed, or unparsable payload | `400` | `invalid_payload` |
| No `text`, `blocks`, or `attachments` | `400` | `no_text` |
| More than the attachment limit | `400` | `too_many_attachments` |
| Body over the configured limit | `413` | `payload_too_large` |
| Non-POST request | `405` | `method_not_allowed` |

Empty and whitespace-only bodies are reported as `invalid_payload`.

State-dependent Slack errors are not emitted automatically because the generic
receiver cannot know your tenant, token, team, channel, or policy state. Return
one of the exported response errors, or construct your own:

```go
return hooksink.ErrInvalidToken
return hooksink.NewResponseError(http.StatusForbidden, "action_prohibited")
```

## Message Shape

The public top-level type is stable:

```go
type Message struct {
	Text         string
	Blocks       slack.Blocks
	Attachments  []slack.Attachment
	Username     string
	IconEmoji    string
	IconURL      string
	Channel      string
	Mrkdwn       *bool
	ResponseType string
	Extra        map[string]json.RawMessage
}
```

Unknown top-level JSON keys are preserved in `Extra`. `Mrkdwn` is a pointer so
`{"mrkdwn":false}` is distinguishable from an absent field.

## Non-Goals

- Sending messages to Slack.
- Mocking Slack Web API methods such as `chat.postMessage`.
- Slack request signing verification for Events API requests.
- OAuth, scopes, token validation, channel binding, persistence, retries, or
  delivery guarantees.
