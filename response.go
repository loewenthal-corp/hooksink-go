package hooksink

import (
	"errors"
	"io"
	"net/http"
	"strings"
)

const textPlain = "text/plain; charset=utf-8"

func writeOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", textPlain)
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

func writeError(w http.ResponseWriter, err error, strict bool) {
	if errors.Is(err, ErrMethodNotAllowed) {
		w.Header().Set("Allow", http.MethodPost)
	}

	status, body := responseStatusAndBody(err, strict)
	w.Header().Set("Content-Type", textPlain)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func responseStatusAndBody(err error, strict bool) (int, string) {
	respErr := errorResponse(err)
	if strict {
		return responseStatus(respErr), respErr.Error()
	}

	switch {
	case errors.Is(err, ErrInvalidPayload), errors.Is(err, ErrNoText), errors.Is(err, ErrTooManyAttachments):
		return http.StatusBadRequest, "bad request"
	case respErr.Status >= http.StatusInternalServerError:
		return responseStatus(respErr), "internal server error"
	default:
		return responseStatus(respErr), strings.ToLower(http.StatusText(responseStatus(respErr)))
	}
}

func errorResponse(err error) *ResponseError {
	var respErr *ResponseError
	if errors.As(err, &respErr) && respErr != nil {
		return respErr
	}
	return ErrInternal
}

func responseStatus(err *ResponseError) int {
	if err == nil || err.Status == 0 {
		return http.StatusInternalServerError
	}
	return err.Status
}
