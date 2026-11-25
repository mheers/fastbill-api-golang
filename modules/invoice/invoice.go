// Package invoice wraps the invoice.* services: outgoing invoices, drafts
// and credit notes.
package invoice

import (
	"context"

	"github.com/mheers/fastbill-api-golang"
)

// Client calls the invoice services.
type Client struct {
	r fastbill.Requester
}

// NewClient returns an invoice client.
func NewClient(r fastbill.Requester) *Client {
	return &Client{r: r}
}

// Get lists invoices. A nil filter lists all (FastBill may leave drafts out
// then; filter by Type to be sure).
func (c *Client) Get(ctx context.Context, page fastbill.Page, filter *Filter) ([]Invoice, error) {
	var res getResponse
	if err := c.r.Do(ctx, fastbill.GetRequest("invoice.get", page, filter), &res); err != nil {
		return nil, err
	}
	return res.Invoices, nil
}

// Create creates a draft invoice.
func (c *Client) Create(ctx context.Context, req *Request) (CreateResponse, error) {
	var res CreateResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.create", req), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.create")
}

// Update changes a draft invoice. It returns the invoice and document IDs
// FastBill reports for the update.
func (c *Client) Update(ctx context.Context, req *Request) (UpdateResponse, error) {
	var res UpdateResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.update", req), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.update")
}

// Delete deletes a draft invoice.
func (c *Client) Delete(ctx context.Context, invoiceID fastbill.ID) error {
	return c.status(ctx, "invoice.delete", idRequest{InvoiceID: invoiceID})
}

// Complete finalizes a draft; FastBill assigns the invoice number.
func (c *Client) Complete(ctx context.Context, invoiceID fastbill.ID) (CompleteResponse, error) {
	var res CompleteResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.complete", idRequest{InvoiceID: invoiceID}), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.complete")
}

// Cancel cancels an invoice completely. It returns the ID of the
// cancellation invoice FastBill wrote.
func (c *Client) Cancel(ctx context.Context, invoiceID fastbill.ID) (CancelResponse, error) {
	return c.cancel(ctx, cancelRequest{InvoiceID: invoiceID})
}

// CancelPartial cancels part of an invoice. After a partial cancellation,
// only further partial cancellations are possible.
func (c *Client) CancelPartial(ctx context.Context, req *PartialCancellationRequest) (CancelResponse, error) {
	return c.cancel(ctx, cancelRequest{
		InvoiceID:           req.InvoiceID,
		PartialCancellation: fastbill.Yes,
		ItemText:            req.ItemText,
		NetAmount:           req.NetAmount,
		VatPercent:          req.VatPercent,
	})
}

func (c *Client) cancel(ctx context.Context, req cancelRequest) (CancelResponse, error) {
	var res CancelResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.cancel", req), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.cancel")
}

// Lock locks an invoice against changes.
func (c *Client) Lock(ctx context.Context, invoiceID fastbill.ID) error {
	return c.status(ctx, "invoice.lock", idRequest{InvoiceID: invoiceID})
}

// SendByEmail emails the invoice PDF.
func (c *Client) SendByEmail(ctx context.Context, req *SendByEmailRequest) error {
	return c.status(ctx, "invoice.sendbyemail", req)
}

// SendByPost sends the invoice by letter, paid with FastBill credits.
func (c *Client) SendByPost(ctx context.Context, invoiceID fastbill.ID) (SendByPostResponse, error) {
	var res SendByPostResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.sendbypost", idRequest{InvoiceID: invoiceID}), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.sendbypost")
}

// SetPaid marks an invoice as paid.
func (c *Client) SetPaid(ctx context.Context, req *SetPaidRequest) (SetPaidResponse, error) {
	var res SetPaidResponse
	if err := c.r.Do(ctx, fastbill.DataRequest("invoice.setpaid", req), &res); err != nil {
		return res, err
	}
	return res, fastbill.StatusResponse{Status: res.Status}.Err("invoice.setpaid")
}

func (c *Client) status(ctx context.Context, service string, data any) error {
	var res fastbill.StatusResponse
	if err := c.r.Do(ctx, fastbill.DataRequest(service, data), &res); err != nil {
		return err
	}
	return res.Err(service)
}
