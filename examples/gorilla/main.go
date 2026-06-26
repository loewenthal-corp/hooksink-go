package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func main() {
	h := hooksink.New(func(_ context.Context, d *hooksink.Delivery) error {
		log.Printf("from %s: %q (%d blocks)", d.Request.URL.Path, d.Message.Text, len(d.Message.Blocks.BlockSet))
		return nil
	})

	r := mux.NewRouter()
	r.Handle("/hook/{id}", h).Methods(http.MethodPost)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
