package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func main() {
	h := hooksink.New(func(_ context.Context, d *hooksink.Delivery) error {
		log.Printf("from %s: %q (%d blocks)", d.Request.URL.Path, d.Message.Text, len(d.Message.Blocks.BlockSet))
		return nil
	})

	g := gin.Default()
	g.POST("/hook/:id", gin.WrapH(h))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           g,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
