package main

import (
	"context"
	"log"
	"net/http"
	"time"

	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func main() {
	h := hooksink.New(func(_ context.Context, d *hooksink.Delivery) error {
		log.Printf("from %s: %q (%d blocks)", d.Request.URL.Path, d.Message.Text, len(d.Message.Blocks.BlockSet))
		return nil
	})

	mux := http.NewServeMux()
	mux.Handle("POST /services/{team}/{bot}/{token}", h)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
