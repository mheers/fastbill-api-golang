package fastbill_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mheers/fastbill-api-golang"
	"github.com/mheers/fastbill-api-golang/internal/fastbilltest"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("body failed") }
func (failingBody) Close() error             { return nil }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDoSendsAddOnUserHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email, key, ok := r.BasicAuth(); !ok || email != "addon@example.com" || key != "api-key" {
			t.Errorf("basic auth = %q, %q, %v", email, key, ok)
		}
		if got := r.Header.Get("X-Username"); got != "user@example.com" {
			t.Errorf("X-Username = %q", got)
		}
		if got := r.Header.Get("X-Password"); got != "secret" {
			t.Errorf("X-Password = %q", got)
		}
		_, _ = io.WriteString(w, `{"RESPONSE":{"STATUS":"success"}}`)
	}))
	defer server.Close()
	client := fastbill.NewClient("addon@example.com", "api-key",
		fastbill.WithBaseURL(server.URL),
		fastbill.WithUserAuth("user@example.com", "secret"),
	)
	if err := client.Do(ctx, fastbill.DataRequest("customer.get", nil), nil); err != nil {
		t.Fatal(err)
	}
}

func TestErrorStrings(t *testing.T) {
	apiErr := &fastbill.APIError{Service: "invoice.get", Messages: []string{"one", "two"}}
	if got, want := apiErr.Error(), "fastbill invoice.get: one; two"; got != want {
		t.Errorf("APIError = %q, want %q", got, want)
	}
	statusErr := &fastbill.StatusError{StatusCode: http.StatusServiceUnavailable, Body: "down"}
	if got, want := statusErr.Error(), "fastbill: HTTP 503: down"; got != want {
		t.Errorf("StatusError = %q, want %q", got, want)
	}
}

func TestDoEncodeErrors(t *testing.T) {
	client := fastbill.NewClient("a", "b")
	if err := client.Do(ctx, fastbill.Request{Service: "x", Data: make(chan int)}, nil); err == nil || !strings.Contains(err.Error(), "encode request") {
		t.Errorf("Do = %v", err)
	}
	if err := client.DoMultipart(ctx, fastbill.Request{Service: "x", Data: make(chan int)}, strings.NewReader("x"), "x.pdf", nil); err == nil || !strings.Contains(err.Error(), "encode request") {
		t.Errorf("DoMultipart = %v", err)
	}
}

func TestDoMultipartReadError(t *testing.T) {
	client := fastbill.NewClient("a", "b")
	err := client.DoMultipart(ctx, fastbill.Request{Service: "document.create"}, failingReader{}, "x.pdf", nil)
	if err == nil || !strings.Contains(err.Error(), "read file") {
		t.Errorf("DoMultipart = %v", err)
	}
}

func TestDoInvalidBaseURL(t *testing.T) {
	client := fastbill.NewClient("a", "b", fastbill.WithBaseURL("://bad"))
	if err := client.Do(ctx, fastbill.Request{Service: "x"}, nil); err == nil {
		t.Error("Do with an invalid base URL passed")
	}
}

func TestDoDecodeErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		body string
		out  any
		want string
	}{
		"envelope": {body: `not json`, want: "decode response"},
		"response": {body: `{"RESPONSE":{"N":"abc"}}`, out: &struct {
			N int `json:"N"`
		}{}, want: "decode response"},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			client := fastbill.NewClient("a", "b", fastbill.WithBaseURL(server.URL))
			err := client.Do(ctx, fastbill.Request{Service: "x"}, tc.out)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestDoStatusErrorTruncatesBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("x", 400), http.StatusInternalServerError)
	}))
	defer server.Close()
	err := fastbill.NewClient("a", "b", fastbill.WithBaseURL(server.URL)).Do(ctx, fastbill.Request{Service: "x"}, nil)
	var statusErr *fastbill.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("err = %v", err)
	}
	if len(statusErr.Body) != 303 || !strings.HasSuffix(statusErr.Body, "…") {
		t.Errorf("body = %q (len %d)", statusErr.Body, len(statusErr.Body))
	}
}

func TestDoResponseTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", (20<<20)+1))
	}))
	defer server.Close()
	err := fastbill.NewClient("a", "b", fastbill.WithBaseURL(server.URL)).Do(ctx, fastbill.Request{Service: "x"}, nil)
	if err == nil || !strings.Contains(err.Error(), "response larger than") {
		t.Errorf("err = %v", err)
	}
}

func TestDoReadErrors(t *testing.T) {
	for name, transport := range map[string]http.RoundTripper{
		"transport": roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("dial failed") }),
		"body": roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: failingBody{}}, nil
		}),
	} {
		t.Run(name, func(t *testing.T) {
			client := fastbill.NewClient("a", "b", fastbill.WithHTTPClient(&http.Client{Transport: transport}))
			if err := client.Do(ctx, fastbill.Request{Service: "x"}, nil); err == nil {
				t.Error("Do passed")
			}
		})
	}
}

func TestDownloadEdgeCases(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/empty":
			w.WriteHeader(http.StatusOK)
		case "/html":
			_, _ = io.WriteString(w, "<html>sign in</html>")
		case "/octet":
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte{0, 1, 2, 3})
		case "/error":
			http.Error(w, "nope", http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	client := fastbill.NewClient("a", "b", fastbill.WithHTTPClient(server.Client()))

	if _, _, err := client.Download(ctx, server.URL+"/empty"); err == nil || !strings.Contains(err.Error(), "empty body") {
		t.Errorf("empty = %v", err)
	}
	if _, _, err := client.Download(ctx, server.URL+"/html"); !errors.Is(err, fastbill.ErrNotDocument) {
		t.Errorf("html = %v", err)
	}
	body, contentType, err := client.Download(ctx, server.URL+"/octet")
	if err != nil || string(body) != "\x00\x01\x02\x03" || contentType != "application/pdf" {
		t.Errorf("octet = %q, %q, %v", body, contentType, err)
	}
	var statusErr *fastbill.StatusError
	if _, _, err := client.Download(ctx, server.URL+"/error"); !errors.As(err, &statusErr) {
		t.Errorf("error = %v", err)
	}

	for _, documentURL := range []string{"http://example.com/x", "https://", "::"} {
		if _, _, err := client.Download(ctx, documentURL); err == nil || !strings.Contains(err.Error(), "not a document URL") {
			t.Errorf("%s: %v", documentURL, err)
		}
	}

	failing := fastbill.NewClient("a", "b", fastbill.WithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("dial failed") }),
	}))
	if _, _, err := failing.Download(ctx, "https://example.com/x"); err == nil {
		t.Error("download with a failing transport passed")
	}
}

func TestRecordsError(t *testing.T) {
	server := fastbilltest.New(t, func(fastbilltest.Received) string { return `{"ERRORS":["boom"]}` })
	if _, err := fastbill.Records(ctx, server.Client(), "invoice.get", "INVOICES", nil, fastbill.Page{}); err == nil {
		t.Error("Records passed")
	}
}

func TestAllErrors(t *testing.T) {
	all, err := fastbill.All(ctx, 2, func(_ context.Context, p fastbill.Page) ([]int, error) {
		if p.Offset == 0 {
			return []int{1, 2}, nil
		}
		return []int{3}, nil
	})
	if err != nil || len(all) != 3 || all[2] != 3 {
		t.Errorf("All = %v, %v", all, err)
	}
	if _, err := fastbill.All(ctx, 0, func(context.Context, fastbill.Page) ([]int, error) { return nil, nil }); err == nil {
		t.Error("page size 0 passed")
	}
	if _, err := fastbill.All(ctx, fastbill.MaxLimit+1, func(context.Context, fastbill.Page) ([]int, error) { return nil, nil }); err == nil {
		t.Error("page size above MaxLimit passed")
	}
	if _, err := fastbill.All(ctx, 2, func(context.Context, fastbill.Page) ([]int, error) {
		return nil, errors.New("boom")
	}); err == nil {
		t.Error("get error passed")
	}
}

func TestTypeStringsAndErrors(t *testing.T) {
	if fastbill.ID("7").String() != "7" {
		t.Error("ID.String")
	}
	if fastbill.Number("1.5").String() != "1.5" {
		t.Error("Number.String")
	}
	if fastbill.NewFlag(false) != fastbill.No {
		t.Error("NewFlag(false)")
	}
	if _, err := fastbill.Number("n/a").Float64(); err == nil {
		t.Error("Float64 of a non-number passed")
	}

	var id fastbill.ID
	if err := json.Unmarshal([]byte(`{}`), &id); err == nil {
		t.Error("ID from an object passed")
	}
	var number fastbill.Number
	if err := json.Unmarshal([]byte(`[]`), &number); err == nil {
		t.Error("Number from a list passed")
	}
	var flag fastbill.Flag
	if err := json.Unmarshal([]byte(`{}`), &flag); err == nil {
		t.Error("Flag from an object passed")
	}

	var ints fastbill.List[int]
	for _, input := range []string{`["x"]`, `{"a":"x"}`, `"x"`, `123`} {
		if err := json.Unmarshal([]byte(input), &ints); err == nil {
			t.Errorf("List[int] from %s passed", input)
		}
	}

	var byName fastbill.List[struct{ N string }]
	if err := json.Unmarshal([]byte(`{"b":{"N":"b"},"a":{"N":"a"}}`), &byName); err != nil {
		t.Fatal(err)
	}
	if len(byName) != 2 || byName[0].N != "a" || byName[1].N != "b" {
		t.Errorf("byName = %+v", byName)
	}

	out, err := json.Marshal(struct {
		N fastbill.Number `json:"N"`
	}{})
	if err != nil || string(out) != `{"N":null}` {
		t.Errorf("empty Number = %s, %v", out, err)
	}
}
