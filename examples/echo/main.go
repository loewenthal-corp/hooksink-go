package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	hooksink "github.com/loewenthal-corp/hooksink-go"
)

func main() {
	h := hooksink.New(func(_ context.Context, d *hooksink.Delivery) error {
		log.Printf("from %s: %q (%d blocks)", d.Request.URL.Path, d.Message.Text, len(d.Message.Blocks.BlockSet))
		return nil
	})

	e := echo.New()
	e.POST("/hook/:id", echo.WrapHandler(h))
	e.Server.ReadHeaderTimeout = 5 * time.Second

	log.Fatal(e.Start(":8080"))
}
