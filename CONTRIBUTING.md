# Contributing

Thanks for taking the time to improve the client. Please open an issue before
a larger change so the direction is agreed first.

## Development setup

You need Go 1.25 or newer. The client has no third-party dependencies.

```bash
make test     # go test ./...
make race     # go test -race ./...
make cover    # coverage gate: 90% per library package
make check    # gofmt, vet, tests and the coverage gate
```

Static analysis in CI runs `go vet`, `staticcheck` and `govulncheck`.

## Tests come first

This repository uses test-driven development:

1. Write a failing test for the behaviour or field you are adding.
2. Run it and confirm it fails for the intended reason.
3. Add the smallest implementation that makes it pass.
4. Refactor with the test in place.

A change without a test that pins it is not ready to merge. Tests must assert
on what ships: the outgoing JSON, the decoded struct, the returned error. A
test that only executes code, or that re-implements the function and compares
the two, is decoration.

Every test names its unhappy paths as well as the success path. For a client
method that usually means at least one test where FastBill returns
`RESPONSE.ERRORS` and one where it reports a failed `STATUS`.

## Coverage

The gate is **90% per library package** (statements). New and changed code
must carry 90% of its own lines. `examples/` and the `internal/fastbilltest`
helper are excluded because they are not library code.

Run the gate with:

```bash
make cover
```

## Style

- `gofmt` is mandatory; CI fails on unformatted files.
- Keep comments in the voice of the existing code: explain why, not what.
- Public identifiers need doc comments that start with the identifier name.
- Do not add a dependency to the module without a strong reason. The client
  is deliberately standard-library only.
- Field names follow FastBill's JSON names; document allowed values and
  required fields in the Go doc comment.

## Commit and pull request

Use a conventional commit subject (`feat:`, `fix:`, `test:`, `docs:`,
`chore:`). Keep each commit focused and make sure `make check` passes.

In the pull request, describe the API action or field you touched, link the
FastBill documentation, and list the tests that fail without the change.
