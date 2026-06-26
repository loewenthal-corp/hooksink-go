package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func main() {
	h := hooksink.New(func(_ context.Context, d *hooksink.Delivery) error {
		log.Printf("from %s: %q (%d blocks)", d.Request.URL.Path, d.Message.Text, len(d.Message.Blocks.BlockSet))
		return nil
	})

	r := chi.NewRouter()
	r.Method(http.MethodPost, "/hook/{id}", h)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
