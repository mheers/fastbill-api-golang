# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-11-25

First release of the fork.

### Added

- `estimate.Update` for `estimate.update`, including `STATE`, `DUE_DATE` and
  `DELETE_EXISTING_ITEMS`.
- `invoice.CancelPartial` for partial cancellations; `invoice.Cancel` now
  returns the cancellation invoice ID.
- Document and item discount fields on invoices, estimates, recurring invoices
  and items, plus `DISCOUNT_TOTAL` and `IS_GROSS`.
- `REFERENCED_INVOICE_ID`, comments and invoice-chain IDs on invoices and
  recurring invoices.
- Accounting references (`CUSTOMER_ACCOUNTING_NOTE`, `CONTRACT_REFERENCE`,
  `CUSTOMER_ORDER_REFERENCE`, `ORDER_REFERENCE`) on invoices, revenues and
  recurring invoices.
- `CC`/`BCC` on customers, `ARTICLE_ID` on time entries, `VAT_CASE` on
  expenses, `PROJECT_ID` on revenue requests.
- `WithUserAuth` for add-on authentication and `Request.TraceID` for
  `x-trace-id`.
- `cancellation.created` and `cancellation.canceled` webhook events.
- Response IDs from update actions: `invoice.update`, `recurring.update`,
  `contact.update`, `article.update`, `project.update`, `time.update`,
  `customer.update`, `estimate.update`.

### Changed

- Module path is `github.com/mheers/fastbill-api-golang`.
- `estimate.State*` constants use the documented response words; the letter
  codes accepted by `estimate.update` are available as `StateCode*`.
- Update methods that return data now return a response struct instead of only
  an error. This is a breaking change relative to upstream v2.

### Fixed

- `estimate.get` no longer drops `DISCOUNT_*` and `IS_GROSS`.
- `time.get` no longer drops `ARTICLE_ID`.
- `expense.get` no longer drops `VAT_CASE`.

### Tests

- 90% coverage gate per library package; every package is at 94% or above.
- Tests for API errors, transport failures, malformed responses, downloads,
  pagination and the scalar JSON types.

[Unreleased]: https://github.com/mheers/fastbill-api-golang/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mheers/fastbill-api-golang/releases/tag/v0.1.0
