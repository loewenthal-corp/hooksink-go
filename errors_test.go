package hooksink_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func TestResponseError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("wrapped: %w", hooksink.ErrInvalidPayload)
	if !errors.Is(err, hooksink.ErrInvalidPayload) {
		t.Fatal("wrapped ErrInvalidPayload does not match errors.Is")
	}

	custom := hooksink.NewResponseError(http.StatusTeapot, "teapot")
	if custom.Error() != "teapot" {
		t.Fatalf("Error() = %q, want teapot", custom.Error())
	}
	if !errors.Is(custom, hooksink.NewResponseError(http.StatusTeapot, "teapot")) {
		t.Fatal("custom response error did not match equivalent response error")
	}
}
