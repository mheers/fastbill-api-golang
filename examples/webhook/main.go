// Command webhook receives FastBill notifications and logs their type.
//
//	go run ./examples/webhook
//
// Register the endpoint (for example https://example.com/fastbill/webhook)
// with webhook.create. FastBill documents no signature for notifications, so
// use an unguessable URL and treat the payload as untrusted input.
package main

import (
	"log"
	"net/http"

	"github.com/mheers/fastbill-api-golang/modules/webhook"
)

func main() {
	http.HandleFunc("/fastbill/webhook", func(w http.ResponseWriter, r *http.Request) {
		event, err := webhook.ParseEvent(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("FastBill %s: invoice=%v customer=%v", event.Type, event.Invoice, event.Customer)
		w.WriteHeader(http.StatusNoContent)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
