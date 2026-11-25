package estimate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mheers/fastbill-api-golang"
	"github.com/mheers/fastbill-api-golang/internal/fastbilltest"
	"github.com/mheers/fastbill-api-golang/modules/estimate"
)

var ctx = context.Background()

// An estimate as FastBill sends it: IDs and text as strings, amounts as
// numbers, an empty list as {}.
const estimateJSON = `{
	"ESTIMATE_ID": "205", "STATE": "accepted", "CUSTOMER_ID": "59", "CUSTOMER_NUMBER": "1042", "PROJECT_ID": "7",
	"ORGANIZATION": "Kunde AG", "PAYMENT_TYPE": 1, "CURRENCY_CODE": "CHF", "BASE_CURRENCY_CODE": "EUR", "EXCHANGE_RATE": 0.9387,
	"TEMPLATE_ID": "3", "ESTIMATE_NUMBER": "A-2026-05", "ESTIMATE_DATE": "2026-09-01", "DUE_DATE": "2026-09-29",
	"DISCOUNT_BASE_AMOUNT": 1300, "DISCOUNT_PERCENT": 5, "DISCOUNT_AMOUNT": 65, "DISCOUNT_TOTAL": 130, "IS_GROSS": "0",
	"SUB_TOTAL": 1234.56, "VAT_TOTAL": 0, "TOTAL": 1234.56,
	"VAT_ITEMS": {},
	"ITEMS": [{"ESTIMATE_ITEM_ID": 9, "DESCRIPTION": "Workshop", "QUANTITY": "2", "UNIT_PRICE": 617.28, "DISCOUNT_BASE_AMOUNT": 650, "DISCOUNT_PERCENT": 5, "DISCOUNT_AMOUNT": 32.5, "VAT_PERCENT": 0, "COMPLETE_NET": 1234.56, "SORT_ORDER": 1}],
	"DOCUMENT_URL": "https://my.fastbill.com/download/a.pdf"
}`

func TestGet(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string { return `{"ESTIMATES":[` + estimateJSON + `]}` })
	estimates, err := estimate.NewClient(server.Client()).Get(ctx, fastbill.Page{Limit: 100}, &estimate.Filter{CustomerID: "59", StartDate: "2026-09-01"})
	if err != nil {
		t.Fatal(err)
	}
	if len(estimates) != 1 {
		t.Fatalf("estimates = %+v", estimates)
	}
	e := estimates[0]
	if e.EstimateID != "205" || e.State != estimate.StateAccepted || e.PaymentType != "1" || e.ExchangeRate != "0.9387" ||
		e.Total != "1234.56" || e.EstimateNumber != "A-2026-05" {
		t.Errorf("estimate = %+v", e)
	}
	if e.DiscountBaseAmount != "1300" || e.DiscountPercent != "5" || e.DiscountAmount != "65" || e.DiscountTotal != "130" || e.IsGross.Bool() {
		t.Errorf("discounts = %+v", e)
	}
	if len(e.VatItems) != 0 || e.Items[0].EstimateItemID != "9" || e.Items[0].Quantity != "2" || e.Items[0].UnitPrice != "617.28" ||
		e.Items[0].DiscountBaseAmount != "650" || e.Items[0].DiscountPercent != "5" || e.Items[0].DiscountAmount != "32.5" {
		t.Errorf("vat items = %v, items = %+v", e.VatItems, e.Items)
	}
	got := server.Last(t)
	if got.Service != "estimate.get" || got.Limit != 100 {
		t.Errorf("request = %+v", got)
	}
	fastbilltest.JSONEqual(t, got.Filter, `{"CUSTOMER_ID":"59","START_DATE":"2026-09-01"}`)
}

func TestCreate(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","ESTIMATE_ID":206,"INVOICE_NUMBER":"A-2026-06","DOCUMENT_ID":91}`
	})
	res, err := estimate.NewClient(server.Client()).Create(ctx, &estimate.Request{
		CustomerID:      "59",
		EstimateDate:    "2026-09-01",
		DiscountPercent: fastbill.NewInt(5),
		DiscountAmount:  fastbill.NewNumber(65.5),
		IsGross:         fastbill.No,
		Items:           []estimate.Item{{Description: "Workshop", Quantity: fastbill.NewInt(2), UnitPrice: fastbill.NewNumber(617.28), VatPercent: fastbill.NewInt(0)}},
	})
	if err != nil || res.EstimateID != "206" || res.InvoiceNumber != "A-2026-06" || res.DocumentID != "91" {
		t.Fatalf("Create = %+v, %v", res, err)
	}
	got := server.Last(t)
	if got.Service != "estimate.create" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{
		"CUSTOMER_ID": "59", "ESTIMATE_DATE": "2026-09-01",
		"DISCOUNT_PERCENT": 5, "DISCOUNT_AMOUNT": 65.5, "IS_GROSS": "0",
		"ITEMS": [{"DESCRIPTION": "Workshop", "QUANTITY": 2, "UNIT_PRICE": 617.28, "VAT_PERCENT": 0}]
	}`)
}

func TestUpdate(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string {
		return `{"STATUS":"success","ESTIMATE_ID":205}`
	})
	res, err := estimate.NewClient(server.Client()).Update(ctx, &estimate.UpdateRequest{
		EstimateID:          "205",
		State:               estimate.StateSent,
		DueDate:             "2026-10-15",
		DeleteExistingItems: fastbill.Yes,
		Request:             estimate.Request{DiscountPercent: fastbill.NewInt(10)},
	})
	if err != nil || res.EstimateID != "205" {
		t.Fatalf("Update = %+v, %v", res, err)
	}
	got := server.Last(t)
	if got.Service != "estimate.update" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{
		"ESTIMATE_ID":"205","STATE":"sent","DUE_DATE":"2026-10-15",
		"DELETE_EXISTING_ITEMS":"1","DISCOUNT_PERCENT":10
	}`)
}

func TestOtherServices(t *testing.T) {
	server := fastbilltest.New(t, func(r fastbilltest.Received) string {
		if r.Service == "estimate.createinvoice" {
			return `{"STATUS":"success","INVOICE_ID":3150,"DOCUMENT_ID":92,"INVOICE_NUMBER":"00280"}`
		}
		return `{"STATUS":"success"}`
	})
	c := estimate.NewClient(server.Client())

	if res, err := c.CreateInvoice(ctx, "205"); err != nil || res.InvoiceID != "3150" || res.DocumentID != "92" || res.InvoiceNumber != "00280" {
		t.Errorf("CreateInvoice = %+v, %v", res, err)
	}
	got := server.Last(t)
	if got.Service != "estimate.createinvoice" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{"ESTIMATE_ID":"205"}`)

	if err := c.Delete(ctx, "205"); err != nil {
		t.Error(err)
	}
	got = server.Last(t)
	if got.Service != "estimate.delete" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{"ESTIMATE_ID":"205"}`)

	if err := c.SendByEmail(ctx, &estimate.SendByEmailRequest{
		EstimateID: "205", Recipient: estimate.Recipient{To: "a@example.com", Cc: "b@example.com"},
		Subject: "Angebot", ReceiptConfirmation: fastbill.Yes,
	}); err != nil {
		t.Error(err)
	}
	got = server.Last(t)
	if got.Service != "estimate.sendbyemail" {
		t.Errorf("service = %s", got.Service)
	}
	fastbilltest.JSONEqual(t, got.Data, `{
		"ESTIMATE_ID": "205", "RECIPIENT": {"TO": "a@example.com", "CC": "b@example.com"},
		"SUBJECT": "Angebot", "RECEIPT_CONFIRMATION": "1"
	}`)
}

func TestErrors(t *testing.T) {
	server := fastbilltest.New(t, func(r fastbilltest.Received) string {
		switch r.Service {
		case "estimate.delete":
			return `{"STATUS":"failed"}`
		case "estimate.createinvoice":
			return `{"STATUS":"failed","INVOICE_ID":0}`
		}
		return `{"ERRORS":["Estimate not found"]}`
	})
	c := estimate.NewClient(server.Client())
	var apiErr *fastbill.APIError
	if _, err := c.Get(ctx, fastbill.Page{}, nil); !errors.As(err, &apiErr) || apiErr.Messages[0] != "Estimate not found" {
		t.Errorf("Get: %v", err)
	}
	if err := c.Delete(ctx, "1"); !errors.As(err, &apiErr) {
		t.Errorf("Delete with failed status: %v", err)
	}
	if _, err := c.CreateInvoice(ctx, "1"); !errors.As(err, &apiErr) {
		t.Errorf("CreateInvoice with failed status: %v", err)
	}
	if err := c.SendByEmail(ctx, &estimate.SendByEmailRequest{EstimateID: "1"}); !errors.As(err, &apiErr) {
		t.Errorf("SendByEmail: %v", err)
	}
	if _, err := c.Update(ctx, &estimate.UpdateRequest{EstimateID: "1"}); !errors.As(err, &apiErr) {
		t.Errorf("Update with failed status: %v", err)
	}
}
