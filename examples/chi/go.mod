module github.com/loewenthal-corp/hooksink-go/examples/chi

go 1.22

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/loewenthal-corp/hooksink-go v0.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/slack-go/slack v0.17.3 // indirect
)

replace github.com/loewenthal-corp/hooksink-go => ../..
