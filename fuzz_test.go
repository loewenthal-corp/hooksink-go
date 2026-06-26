package hooksink_test

import (
	"encoding/json"
	"errors"
	"testing"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func FuzzParseBody(f *testing.F) {
	seeds := []struct {
		contentType string
		body        []byte
	}{
		{contentType: "application/json", body: []byte(`{"text":"hello"}`)},
		{contentType: "application/json; charset=utf-8", body: []byte(`{"blocks":[{"type":"divider"}]}`)},
		{contentType: "application/x-www-form-urlencoded", body: []byte(`payload=%7B%22text%22%3A%22hello%22%7D`)},
		{contentType: "", body: []byte(`{"attachments":[{"text":"hello"}]}`)},
		{contentType: "application/json", body: []byte(`{"text":`)},
		{contentType: "application/json", body: nil},
	}

	for _, seed := range seeds {
		f.Add(seed.contentType, seed.body)
	}

	f.Fuzz(func(t *testing.T, contentType string, body []byte) {
		msg, err := hooksink.ParseBody(contentType, body)
		if err != nil {
			var responseErr *hooksink.ResponseError
			if !errors.As(err, &responseErr) {
				t.Fatalf("ParseBody returned non-response error: %T %v", err, err)
			}
			return
		}

		raw, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal parsed message: %v", err)
		}
		if !json.Valid(raw) {
			t.Fatalf("marshaled message is invalid JSON: %q", raw)
		}
	})
}
