package hooksink

import "net/http"

// ResponseError maps an application or parse failure to Slack-compatible plain
// text response semantics.
type ResponseError struct {
	Status int
	Code   string
}

// NewResponseError constructs a ResponseError with the supplied HTTP status and
// plain-text response code.
func NewResponseError(status int, code string) *ResponseError {
	return &ResponseError{Status: status, Code: code}
}

// Error returns the plain-text response code.
func (e *ResponseError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != "" {
		return e.Code
	}
	if text := http.StatusText(e.Status); text != "" {
		return text
	}
	return "response_error"
}

// Is reports whether two response errors carry the same status and code. A
// target with a zero Status or empty Code treats that field as a wildcard.
func (e *ResponseError) Is(target error) bool {
	t, ok := target.(*ResponseError)
	if !ok || e == nil || t == nil {
		return false
	}
	statusMatches := t.Status == 0 || e.Status == t.Status
	codeMatches := t.Code == "" || e.Code == t.Code
	return statusMatches && codeMatches
}

var (
	// ErrInvalidPayload is returned for empty bodies, malformed JSON, and other
	// unparsable payloads.
	ErrInvalidPayload = &ResponseError{Status: http.StatusBadRequest, Code: "invalid_payload"}
	// ErrNoText is returned when text, blocks, and attachments are all absent or
	// empty.
	ErrNoText = &ResponseError{Status: http.StatusBadRequest, Code: "no_text"}
	// ErrTooManyAttachments is returned when a payload exceeds the configured
	// attachment count.
	ErrTooManyAttachments = &ResponseError{Status: http.StatusBadRequest, Code: "too_many_attachments"}
	// ErrPayloadTooLarge is returned when the request body exceeds the configured
	// size limit.
	ErrPayloadTooLarge = &ResponseError{Status: http.StatusRequestEntityTooLarge, Code: "payload_too_large"}
	// ErrMethodNotAllowed is returned for non-POST requests.
	ErrMethodNotAllowed = &ResponseError{Status: http.StatusMethodNotAllowed, Code: "method_not_allowed"}
	// ErrInternal is used for unexpected callback or handler failures.
	ErrInternal = &ResponseError{Status: http.StatusInternalServerError, Code: "rollup_error"}

	// ErrChannelIsArchived can be returned by consumers that model Slack channel
	// state.
	ErrChannelIsArchived = &ResponseError{Status: http.StatusGone, Code: "channel_is_archived"}
	// ErrNoTeam can be returned by consumers that model team lookup state.
	ErrNoTeam = &ResponseError{Status: http.StatusNotFound, Code: "no_team"}
	// ErrNoService can be returned by consumers that model service lookup state.
	ErrNoService = &ResponseError{Status: http.StatusNotFound, Code: "no_service"}
	// ErrNoServiceID can be returned by consumers that model service ID lookup
	// state.
	ErrNoServiceID = &ResponseError{Status: http.StatusNotFound, Code: "no_service_id"}
	// ErrTeamDisabled can be returned by consumers that model disabled teams.
	ErrTeamDisabled = &ResponseError{Status: http.StatusForbidden, Code: "team_disabled"}
	// ErrActionProhibited can be returned by consumers enforcing posting policy.
	ErrActionProhibited = &ResponseError{Status: http.StatusForbidden, Code: "action_prohibited"}
	// ErrPostingToGeneralChannelDenied can be returned by consumers enforcing
	// general-channel posting policy.
	ErrPostingToGeneralChannelDenied = &ResponseError{
		Status: http.StatusForbidden,
		Code:   "posting_to_general_channel_denied",
	}
	// ErrInvalidToken can be returned by consumers enforcing per-URL tokens.
	ErrInvalidToken = &ResponseError{Status: http.StatusForbidden, Code: "invalid_token"}
	// ErrNoActiveHooks can be returned by consumers that model inactive hooks.
	ErrNoActiveHooks = &ResponseError{Status: http.StatusNotFound, Code: "no_active_hooks"}
)
