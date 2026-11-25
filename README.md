# fastbill-api-golang

[![CI](https://github.com/mheers/fastbill-api-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/mheers/fastbill-api-golang/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mheers/fastbill-api-golang.svg)](https://pkg.go.dev/github.com/mheers/fastbill-api-golang)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Go client for the [FastBill API](https://apidocs.fastbill.com/fastbill/en/fundamentals.html):
customers, invoices, estimates, expenses, revenues, products, projects, work
times, documents, templates and webhooks.

The client uses FastBill's JSON format, standard library only, and every call
takes a `context.Context`.

This repository is a fork of
[Malsch-Solutions/fastbill-go-sdk](https://github.com/Malsch-Solutions/fastbill-go-sdk)
v2 (MIT). It keeps that module layout and adds the current API surface, a
coverage gate and the fixes listed in [Changes from upstream](#changes-from-upstream).
See [NOTICE](NOTICE) for the attribution.

## Installation

Requires Go 1.25 or newer.

```bash
go get github.com/mheers/fastbill-api-golang
```

## Quick start

The client authenticates with the email address of your FastBill user and the
API key from your FastBill account settings.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mheers/fastbill-api-golang"
	"github.com/mheers/fastbill-api-golang/modules/invoice"
)

func main() {
	ctx := context.Background()
	client := fastbill.NewClient(os.Getenv("FASTBILL_EMAIL"), os.Getenv("FASTBILL_API_KEY"))
	invoices := invoice.NewClient(client)

	drafts, err := invoices.Get(ctx, fastbill.Page{Limit: 100}, &invoice.Filter{Type: invoice.TypeDraft})
	if err != nil {
		log.Fatal(err)
	}
	for _, inv := range drafts {
		fmt.Println(inv.InvoiceID, inv.InvoiceDate, inv.Total, inv.CurrencyCode)
	}
}
```

Every service has a typed client under `modules/`; `fastbill.NewClient`
returns the shared transport.

## Authentication

```go
// Own account.
client := fastbill.NewClient(email, apiKey)

// Add-on or app acting for a user: authenticate as the add-on and send the
// user's credentials in X-Username/X-Password.
client = fastbill.NewClient(addonEmail, addonKey,
	fastbill.WithUserAuth(userEmail, userPassword))
```

FastBill's API is stateless: every request carries the credentials.

An optional trace ID helps FastBill support follow a request:

```go
req := fastbill.DataRequest("invoice.get", nil)
req.TraceID = "c3a1e2b6-4f09-4c1b-9d3e-5b8a6f2d9f7c"
err := client.Do(ctx, req, &out)
```

## Services

| Package | Services |
|---|---|
| `modules/article` | `article.get/create/update/delete` |
| `modules/contact` | `contact.get/create/update/delete` |
| `modules/customer` | `customer.get/create/update/delete` |
| `modules/document` | `document.get/create` (upload) |
| `modules/estimate` | `estimate.get/create/update/delete/sendbyemail/createinvoice` |
| `modules/expense` | `expense.get/create` (optional receipt upload) |
| `modules/invoice` | `invoice.get/create/update/delete/complete/cancel/lock/sendbyemail/sendbypost/setpaid` |
| `modules/item` | `item.get/delete` |
| `modules/project` | `project.get/create/update/delete` |
| `modules/recurring` | `recurring.get/create/update/delete` |
| `modules/revenue` | `revenue.get/create/setpaid/delete` |
| `modules/template` | `template.get` |
| `modules/time` | `time.get/create/update/delete` |
| `modules/webhook` | `webhook.get/create/delete`, notification parsing |

## Common patterns

### Paging through a listing

FastBill returns at most 100 records per call and 10 without a limit.
`fastbill.All` walks the offsets:

```go
all, err := fastbill.All(ctx, fastbill.MaxLimit, func(ctx context.Context, page fastbill.Page) ([]invoice.Invoice, error) {
	return invoices.Get(ctx, page, &invoice.Filter{Type: invoice.TypeOutgoing})
})
```

### Raw records

`fastbill.Records` returns each record exactly as FastBill sent it, including
fields the typed structs do not have:

```go
records, err := fastbill.Records(ctx, client, "invoice.get", "INVOICES", nil, fastbill.Page{Limit: 100})
```

### Errors

FastBill reports business errors in `RESPONSE.ERRORS`, often with HTTP 200.
Those become `*fastbill.APIError`; non-2xx answers become
`*fastbill.StatusError`.

```go
var apiErr *fastbill.APIError
if errors.As(err, &apiErr) {
	log.Printf("FastBill %s: %v", apiErr.Service, apiErr.Messages)
}
```

Services that only report a status return an error unless `STATUS` is
`success`.

### Invoice lifecycle

```go
created, err := invoices.Create(ctx, &invoice.Request{
	CustomerID: "59",
	Items: []invoice.Item{{
		Description: "Support", Quantity: fastbill.NewInt(2), UnitPrice: fastbill.NewNumber(80), VatPercent: fastbill.NewInt(19),
	}},
})

completed, err := invoices.Complete(ctx, created.InvoiceID) // assigns the invoice number
err = invoices.SendByEmail(ctx, &invoice.SendByEmailRequest{
	InvoiceID: created.InvoiceID,
	Recipient: invoice.Recipient{To: "customer@example.com"},
})

// Full cancellation; CancelPartial sends PARTIAL_CANCELLATION with an amount.
cancellation, err := invoices.Cancel(ctx, created.InvoiceID)
fmt.Println(cancellation.InvoiceID) // the cancellation invoice
```

### Documents

`Download` fetches a `DOCUMENT_URL` (invoice PDF or receipt). The URL carries
its own token, so no credentials are sent:

```go
body, contentType, err := client.Download(ctx, inv.DocumentURL)
```

### Webhooks

`webhook.ParseEvent` validates and decodes a FastBill notification. Answer
with `400 Bad Request` when it returns an error:

```go
http.HandleFunc("/fastbill/webhook", func(w http.ResponseWriter, r *http.Request) {
	event, err := webhook.ParseEvent(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch event.Type {
	case webhook.InvoiceCreated:
		log.Printf("invoice %s", event.Invoice.InvoiceID)
	}
	w.WriteHeader(http.StatusNoContent)
})
```

FastBill documents no signature for notifications. Use an unguessable
endpoint URL and treat the payload as untrusted input.

### Types

FastBill sends the same field as a JSON string in one response and as a
number in another. The structs use:

| Type | For | Accepts |
|---|---|---|
| `fastbill.ID` | `*_ID` fields | string, number |
| `fastbill.Number` | amounts, quantities, percentages | number, numeric string; keeps the digits as sent, so money is never rounded through `float64`. `Float64()` parses it |
| `fastbill.Flag` | 0/1 fields (`IS_CANCELED`, …) | `"1"`/`"0"`, `1`/`0`, `true`/`false`; `Bool()` reads it |
| `fastbill.List[T]` | lists in responses | array, PHP object keyed by index, `{}` or `null` when empty |

Dates are strings in FastBill's format (`YYYY-MM-DD`; timestamps
`YYYY-MM-DD hh:mm:ss`).

## Rate limits

FastBill limits calls per hour by plan: Solo 50, Solo-Plus 100, Pro 500,
Premium 1000. Webhooks are available in Pro and Premium. The client does not
throttle; space large jobs out yourself.

## Changes from upstream

- `estimate.update` and the documented estimate state values.
- `estimate.State*` constants use the response words (`accepted`, …); the
  letter codes accepted by `estimate.update` are `estimate.StateCode*`.
- Document and item discounts, `DISCOUNT_TOTAL`, `IS_GROSS`, referenced
  invoices and comments on invoices, estimates and recurring invoices.
- `invoice.CancelPartial` for `PARTIAL_CANCELLATION`, and the cancellation
  invoice ID from `invoice.Cancel`.
- Response IDs from `invoice.update`, `recurring.update`, `contact.update`
  (`ADDITIONAL_INFO`) and the other update actions.
- `CC`/`BCC`, `ARTICLE_ID` on time entries, `VAT_CASE` on expenses,
  accounting references and invoice-chain IDs.
- Add-on authentication (`WithUserAuth`) and request trace IDs.
- `cancellation.created`/`cancellation.canceled` webhook events.
- A 90% per-package coverage gate and tests for every service.

## Development

```bash
make test     # go test ./...
make race     # go test -race ./...
make cover    # coverage gate: 90% per library package
make check    # gofmt, vet, tests and the coverage gate
```

`examples/` and the `internal/fastbilltest` helper are excluded from the
coverage gate.

## License

MIT. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

Original work Copyright (c) 2021 Dominik Malsch Solutions.
Fork changes Copyright (c) 2026 Marcel Heers.
