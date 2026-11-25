// Package fastbill is a client for the FastBill API
// (https://apidocs.fastbill.com/fastbill/en/).
//
// A Client sends requests; the packages under modules/ wrap each service
// (customers, invoices, expenses, …) with typed requests and responses:
//
//	client := fastbill.NewClient(os.Getenv("FASTBILL_EMAIL"), os.Getenv("FASTBILL_API_KEY"))
//	customers, err := customer.NewClient(client).Get(ctx, fastbill.Page{Limit: 100}, nil)
//
// Add-ons and apps that act for a FastBill user keep their own credentials
// for basic authentication and pass the user's with WithUserAuth. A request
// may carry a TraceID, which is sent as x-trace-id.
//
// FastBill mixes JSON numbers and strings for the same field. ID and Number
// accept both, and Number keeps the digits as sent. Records returns records
// unparsed, e.g. to archive them, and Download fetches a DOCUMENT_URL.
// Webhook notifications are parsed with ParseEvent in the webhook package.
//
// FastBill limits API calls per hour by plan. The client does not throttle;
// space calls out yourself for large exports.
package fastbill
