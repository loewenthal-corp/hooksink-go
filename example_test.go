package hooksink_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func ExampleParseBody() {
	msg, err := hooksink.ParseBody("application/json", []byte(`{"text":"build passed"}`))
	if err != nil {
		panic(err)
	}

	fmt.Println(msg.Text)
	// Output: build passed
}

func ExampleParse() {
	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"text":"deploy finished"}`))
	req.Header.Set("Content-Type", "application/json")

	msg, err := hooksink.Parse(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(msg.Text)
	// Output: deploy finished
}

func ExampleNew() {
	h := hooksink.New(func(ctx context.Context, d *hooksink.Delivery) error {
		fmt.Printf("%s: %s\n", d.Request.URL.Path, d.Message.Text)
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/services/T/B/X", strings.NewReader(`{"text":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	fmt.Println(rec.Code, rec.Body.String())
	// Output:
	// /services/T/B/X: hello
	// 200 ok
}
