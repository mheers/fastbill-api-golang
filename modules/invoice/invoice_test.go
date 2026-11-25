package invoice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mheers/fastbill-api-golang"
	"github.com/mheers/fastbill-api-golang/internal/fastbilltest"
	"github.com/mheers/fastbill-api-golang/modules/invoice"
)

var ctx = context.Background()

// An invoice as FastBill sends it: IDs and text as strings, amounts as
// numbers, an empty list as {}.
const invoiceJSON = `{
	"INVOICE_ID": "3141", "TYPE": "outgoing", "CUSTOMER_ID": "59", "CUSTOMER_NUMBER": "1042",
	"INVOICE_NUMBER": "00277", "INVOICE_DATE": "2026-09-01", "IS_CANCELED": "0", "IS_GROSS": 0,
	"PAYMENT_TYPE": "1", "SUB_TOTAL": 1234.5, "VAT_TOTAL": 0, "TOTAL": 1234.5, "CURRENCY_CODE": "EUR",
	"DISCOUNT_BASE_AMOUNT": 2000, "DISCOUNT_PERCENT": 5, "DISCOUNT_AMOUNT": 100, "DISCOUNT_TOTAL": 160,
	"REFERENCED_INVOICE_ID": "3000",
	"COMMENTS": [{"DATE": "2026-09-02", "COMMENT": "Checked", "COMMENT_PUBLIC": "1"}],
	"ITEMS": [{"INVOICE_ITEM_ID": 7, "DESCRIPTION": "Support", "QUANTITY": 1.5, "UNIT": "Std.", "UNIT_PRICE": 823, "VAT_PERCENT": 0, "COMPLETE_NET": 1234.5, "DISCOUNT_BASE_AMOUNT": 900, "DISCOUNT_PERCENT": 10, "DISCOUNT_AMOUNT": 90, "CATEGORY": []}],
	"VAT_ITEMS": [{"VAT_PERCENT": 0, "COMPLETE_NET": 1234.5, "VAT_VALUE": 0}],
	"PAYMENTS": {},
	"DOCUMENT_URL": "https://my.fastbill.com/download/x.pdf"
}`

func TestGet(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string { return `{"INVOICES":[` + invoiceJSON + `]}` })
	invoices, err := invoice.NewClient(server.Client()).Get(ctx, fastbill.Page{Limit: 100}, &invoice.Filter{Type: invoice.TypeDraft, State: "issued"})
	if err != nil {
		t.Fatal(err)
	}
	if len(invoices) != 1 {
		t.Fatalf("invoices = %+v", invoices)
	}
	inv := invoices[0]
	if inv.InvoiceID != "3141" || inv.InvoiceNumber != "00277" || inv.Total != "1234.5" || inv.IsCanceled.Bool() || inv.IsGross != fastbill.No {
		t.Errorf("invoice = %+v", inv)
	}
	if inv.DiscountBaseAmount != "2000" || inv.DiscountPercent != "5" || inv.DiscountAmount != "100" || inv.DiscountTotal != "160" {
		t.Errorf("discounts = %+v", inv)
	}
	if inv.ReferencedInvoiceID != "3000" {
		t.Errorf("referenced invoice = %q", inv.ReferencedInvoiceID)
	}
	if len(inv.Comments) != 1 || inv.Comments[0].Comment != "Checked" || !inv.Comments[0].CommentPublic.Bool() {
		t.Errorf("comments = %+v", inv.Comments)
	}
	if item := inv.Items[0]; item.InvoiceItemID != "7" || item.Quantity != "1.5" || item.Unit != "Std." ||
		item.DiscountBaseAmount != "900" || item.DiscountPercent != "10" || item.DiscountAmount != "90" {
		t.Errorf("item = %+v", item)
	}
	if len(inv.Payments) != 0 || inv.VatItems[0].CompleteNet != "1234.5" {
		t.Errorf("payments = %v, vat items = %v", inv.Payments, inv.VatItems)
	}
	got := server.Last(t)
	if got.Service != "invoice.get" || got.Limit != 100 {
		t.Errorf("request = %+v", got)
	}
	fastbilltest.JSONEqual(t, got.Filter, `{"TYPE":"draft","STATE":"issued"}`)
}

func TestCreate(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","INVOICE_ID":3142,"DOCUMENT_ID":88,"INVOICE_NUMBER":"00279"}`
	})
	res, err := invoice.NewClient(server.Client()).Create(ctx, &invoice.Request{
		CustomerID:      "59",
		VatCase:         "small_business_regulation",
		DiscountPercent: fastbill.NewInt(5),
		DiscountAmount:  fastbill.NewInt(100),
		Items: []invoice.Item{{
			Description: "Support", Quantity: fastbill.NewNumber(1.5), UnitPrice: fastbill.NewInt(80), VatPercent: fastbill.NewInt(0),
		}},
	})
	if err != nil || res.InvoiceID != "3142" || res.DocumentID != "88" || res.InvoiceNumber != "00279" {
		t.Fatalf("Create = %+v, %v", res, err)
	}
	fastbilltest.JSONEqual(t, server.Last(t).Data, `{
		"CUSTOMER_ID": "59", "VAT_CASE": "small_business_regulation",
		"DISCOUNT_PERCENT": 5, "DISCOUNT_AMOUNT": 100,
		"ITEMS": [{"DESCRIPTION": "Support", "QUANTITY": 1.5, "UNIT_PRICE": 80, "VAT_PERCENT": 0}]
	}`)
}

func TestUpdateReturnsResponse(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","INVOICE_ID":3142,"DOCUMENT_ID":89}`
	})
	res, err := invoice.NewClient(server.Client()).Update(ctx, &invoice.Request{
		InvoiceID: "3142", DeleteExistingItems: fastbill.Yes, DiscountPercent: fastbill.NewInt(10),
	})
	if err != nil || res.InvoiceID != "3142" || res.DocumentID != "89" {
		t.Fatalf("Update = %+v, %v", res, err)
	}
	fastbilltest.JSONEqual(t, server.Last(t).Data, `{"INVOICE_ID":"3142","DELETE_EXISTING_ITEMS":"1","DISCOUNT_PERCENT":10}`)
}

func TestCancelReturnsResponse(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","INVOICE_ID":4000}`
	})
	res, err := invoice.NewClient(server.Client()).Cancel(ctx, "3142")
	if err != nil || res.InvoiceID != "4000" {
		t.Fatalf("Cancel = %+v, %v", res, err)
	}
	got := server.Last(t)
	if got.Service != "invoice.cancel" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{"INVOICE_ID":"3142"}`)
}

func TestCancelPartial(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","INVOICE_ID":4001}`
	})
	res, err := invoice.NewClient(server.Client()).CancelPartial(ctx, &invoice.PartialCancellationRequest{
		InvoiceID: "3142", ItemText: "Teilrechnung", NetAmount: fastbill.NewNumber(100), VatPercent: fastbill.NewInt(19),
	})
	if err != nil || res.InvoiceID != "4001" {
		t.Fatalf("CancelPartial = %+v, %v", res, err)
	}
	got := server.Last(t)
	if got.Service != "invoice.cancel" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{
		"INVOICE_ID":"3142","PARTIAL_CANCELLATION":"1","ITEM_TEXT":"Teilrechnung",
		"NET_AMOUNT":100,"VAT_PERCENT":19
	}`)
}

func TestServicesWithStatus(t *testing.T) {
	server := fastbilltest.New(t, func(r fastbilltest.Received) string {
		switch r.Service {
		case "invoice.complete":
			return `{"STATUS":"success","INVOICE_NUMBER":"00278","INVOICE_ID":3143,"DOCUMENT_ID":90}`
		case "invoice.sendbypost":
			return `{"STATUS":"success","REMAINING_CREDITS":"12"}`
		case "invoice.setpaid":
			return `{"STATUS":"success","INVOICE_NUMBER":"00278"}`
		}
		return `{"STATUS":"success"}`
	})
	c := invoice.NewClient(server.Client())

	if res, err := c.Complete(ctx, "1"); err != nil || res.InvoiceNumber != "00278" || res.InvoiceID != "3143" || res.DocumentID != "90" {
		t.Errorf("Complete = %+v, %v", res, err)
	}
	for name, call := range map[string]func() error{
		"invoice.delete": func() error { return c.Delete(ctx, "1") },
		"invoice.lock":   func() error { return c.Lock(ctx, "1") },
	} {
		if err := call(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if got := server.Last(t); got.Service != name {
			t.Errorf("service = %s, want %s", got.Service, name)
		}
		fastbilltest.JSONEqual(t, server.Last(t).Data, `{"INVOICE_ID":"1"}`)
	}
	if err := c.SendByEmail(ctx, &invoice.SendByEmailRequest{InvoiceID: "1", Recipient: invoice.Recipient{To: "a@example.com"}}); err != nil {
		t.Error(err)
	}
	fastbilltest.JSONEqual(t, server.Last(t).Data, `{"INVOICE_ID":"1","RECIPIENT":{"TO":"a@example.com"}}`)
	if res, err := c.SendByPost(ctx, "1"); err != nil || res.RemainingCredits != "12" {
		t.Errorf("SendByPost = %+v, %v", res, err)
	}
	if res, err := c.SetPaid(ctx, &invoice.SetPaidRequest{InvoiceID: "1", PaidDate: "2026-09-30"}); err != nil || res.InvoiceNumber != "00278" {
		t.Errorf("SetPaid = %+v, %v", res, err)
	}
	fastbilltest.JSONEqual(t, server.Last(t).Data, `{"INVOICE_ID":"1","PAID_DATE":"2026-09-30"}`)
}

func TestErrors(t *testing.T) {
	server := fastbilltest.New(t, func(r fastbilltest.Received) string {
		if r.Service == "invoice.get" {
			return `{"ERRORS":["Invoice not found"]}`
		}
		return `{"STATUS":"failed"}`
	})
	c := invoice.NewClient(server.Client())
	var apiErr *fastbill.APIError
	if _, err := c.Get(ctx, fastbill.Page{}, nil); !errors.As(err, &apiErr) || apiErr.Messages[0] != "Invoice not found" {
		t.Errorf("Get: %v", err)
	}
	for name, call := range map[string]func() error{
		"invoice.update": func() error {
			_, err := c.Update(ctx, &invoice.Request{InvoiceID: "1"})
			return err
		},
		"invoice.delete": func() error { return c.Delete(ctx, "1") },
		"invoice.complete": func() error {
			_, err := c.Complete(ctx, "1")
			return err
		},
		"invoice.cancel": func() error {
			_, err := c.Cancel(ctx, "1")
			return err
		},
		"invoice.cancel partial": func() error {
			_, err := c.CancelPartial(ctx, &invoice.PartialCancellationRequest{InvoiceID: "1"})
			return err
		},
		"invoice.lock":        func() error { return c.Lock(ctx, "1") },
		"invoice.sendbyemail": func() error { return c.SendByEmail(ctx, &invoice.SendByEmailRequest{InvoiceID: "1"}) },
		"invoice.sendbypost": func() error {
			_, err := c.SendByPost(ctx, "1")
			return err
		},
		"invoice.setpaid": func() error {
			_, err := c.SetPaid(ctx, &invoice.SetPaidRequest{InvoiceID: "1"})
			return err
		},
	} {
		if err := call(); !errors.As(err, &apiErr) {
			t.Errorf("%s with failed status: %v", name, err)
		}
	}
}

func TestAPIFailures(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string { return `{"ERRORS":["boom"]}` })
	c := invoice.NewClient(server.Client())
	var apiErr *fastbill.APIError
	for name, call := range map[string]func() error{
		"get":      func() error { _, err := c.Get(ctx, fastbill.Page{}, nil); return err },
		"create":   func() error { _, err := c.Create(ctx, &invoice.Request{}); return err },
		"update":   func() error { _, err := c.Update(ctx, &invoice.Request{}); return err },
		"delete":   func() error { return c.Delete(ctx, "1") },
		"complete": func() error { _, err := c.Complete(ctx, "1"); return err },
		"cancel":   func() error { _, err := c.Cancel(ctx, "1"); return err },
		"cancel partial": func() error {
			_, err := c.CancelPartial(ctx, &invoice.PartialCancellationRequest{InvoiceID: "1"})
			return err
		},
		"lock":        func() error { return c.Lock(ctx, "1") },
		"sendbyemail": func() error { return c.SendByEmail(ctx, &invoice.SendByEmailRequest{InvoiceID: "1"}) },
		"sendbypost":  func() error { _, err := c.SendByPost(ctx, "1"); return err },
		"setpaid":     func() error { _, err := c.SetPaid(ctx, &invoice.SetPaidRequest{InvoiceID: "1"}); return err },
	} {
		if err := call(); !errors.As(err, &apiErr) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
