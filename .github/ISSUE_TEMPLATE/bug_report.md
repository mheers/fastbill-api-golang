---
name: Bug report
about: Report a problem with the client
title: ""
labels: bug
assignees: ""
---

**What happened?**

A clear description of the behaviour and what you expected instead.

**Which service or action?**

For example `invoice.create`, `customer.get`, `webhook.ParseEvent`, or the
transport itself.

**Minimal reproduction**

```go
// The smallest program or test that shows the problem.
```

If possible, include the request body FastBill received and the response
`RESPONSE` object. **Remove your email address, API key and any customer
data first.**

**Environment**

- Go version (`go version`):
- Client version or commit:
- FastBill plan (only relevant for rate limits):

**Additional context**

Anything else that helps: logs, the FastBill documentation page for the
action, or a failing test.
