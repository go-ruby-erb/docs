# Contributing

Contributions are welcome. `go-ruby-erb/erb` is built to a small set of
non-negotiable rules — they are what keep the compiler pure-Go, correct, and
MRI-compatible. Please read these before opening a pull request.

## Hard rules

- **Build from source — no vendoring.** Everything compiles from source. Do not
  reach for prebuilt binaries or vendored blobs as a shortcut; being able to
  compile from source is a guarantee of independence.
- **100% test coverage target, enforced in CI.** New code ships with tests, and
  coverage is a CI gate. Fill the error branches, not just the happy path.
- **All GitHub content in English.** Issues, pull requests, commits, comments,
  and discussions are English-only.
- **Differential testing against MRI.** Correctness is defined by reference Ruby.
  A corpus of templates is compiled through both `ruby` and this library, and the
  **emitted Ruby source** *and* the **rendered output** are compared
  byte-for-byte — not approximated from memory. The reference is MRI 4.0.5.
- **Pure Go, cgo disabled.** The whole point is a single static binary with no C
  toolchain. Code must build with `CGO_ENABLED=0`. If a feature seems to need C,
  it needs a pure-Go path instead.
- **The library compiles; it does not evaluate.** Adding template *rendering*
  (the `eval`) to this module is out of scope — that needs an interpreter and
  belongs in the consumer. See [Why a compiler, not a renderer](why.md).

## Workflow

1. Pick or open an issue describing the change.
2. Work test-first: add the differential / unit tests, then make them pass.
3. Run the full suite with coverage and confirm the gate is green:

    ```sh
    COVERPKG=$(go list ./... | paste -sd, -)
    go test -race -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
    go tool cover -func=cover.out | tail -1   # 100.0%
    ```

4. Open a PR in English, referencing the issue.

## Where things live

The compiler — tag scanner, trim modes, the `Compile` / `Compiler` API, and
`ERB::Util` — is in
[`github.com/go-ruby-erb/erb`](https://github.com/go-ruby-erb/erb). This
documentation site is in
[`github.com/go-ruby-erb/docs`](https://github.com/go-ruby-erb/docs). Start from
the [Usage & API](api.md) page and the [Roadmap](roadmap.md) to find the right
place for your change.
