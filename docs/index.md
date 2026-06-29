# go-ruby-erb documentation

**A pure-Go reimplementation of Ruby's ERB template compiler** — the
deterministic, interpreter-independent core of MRI's `ERB::Compiler`, built with
**zero cgo**.

`go-ruby-erb/erb` turns an ERB template string into the **Ruby source that
renders it**, matching MRI 4.0.5 byte-for-byte, and provides `ERB::Util`'s
`html_escape` / `url_encode` helpers. The module path is
`github.com/go-ruby-erb/erb`.

The module is **standalone and reusable**: any Go program can import it. It is
also the ERB backend for [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby),
where the host evaluates the compiled source against a binding.

!!! success "Status: compiler complete — MRI byte-exact"
    Faithful port of MRI's `lib/erb/compiler.rb`: **all tag kinds**
    (`<% %>`, `<%= %>`, `<%# %>`), the `<%%` / `%%>` **literals**, **every trim
    mode** (`-`, `>`, `<>`, `%` and combinations), **magic-comment** detection
    (incl. the emacs `-*- … -*-` form), and **binary-exact** literal-run encoding
    via Ruby's `String#dump`, plus `ERB::Util`'s `HTMLEscape` / `URLEncode`.
    Validated by a **differential oracle** against the system `ruby` — emitted
    source *and* rendered output compared byte-for-byte — at 100% coverage,
    `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets.

## What it is — and isn't

Compiling a template to Ruby source (tag scanning, trim modes, the `<%% %%>`
literals, magic-comment detection, the `String#dump` text encoding) is fully
**deterministic** and needs **no interpreter**, so it lives here as pure Go. The
final `eval(compiled_src, binding)` that produces the rendered string **does**
need a Ruby interpreter and stays in the consumer (e.g. rbgo).

> **This library compiles; the host evaluates.**

## Quick taste

```go
src, magic, err := erb.Compile("Hello <%= name %>!\n", erb.Options{})
// magic == "#coding:UTF-8"
// src   == the Ruby program that, eval'd against a binding, renders the template
```

```go
erb.HTMLEscape(`<a href="x">it's</a>`)
// &lt;a href=&quot;x&quot;&gt;it&#39;s&lt;/a&gt;
erb.URLEncode("100% & more")
// 100%25%20%26%20more
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`erb`](https://github.com/go-ruby-erb/erb) | the compiler — tag scanner, trim modes, the `Compile` / `Compiler` API, and `ERB::Util` (`HTMLEscape` / `URLEncode`) |
| [`docs`](https://github.com/go-ruby-erb/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-erb.github.io`](https://github.com/go-ruby-erb/go-ruby-erb.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-erb/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** The emitted Ruby source and the rendered output match the
  reference `ERB::Compiler` exactly, not approximately.
- **Standalone & reusable.** No dependency on the Ruby runtime; the dependency
  runs the other way (go-embedded-ruby consumes this).
- **100% test coverage** is the target, enforced as a CI gate.

## Where to go next

- [Why a compiler, not a renderer](why.md) — the deterministic/interpreter split,
  and why the compiler can live in pure Go while `eval` cannot.
- [Usage & API](api.md) — `Compile`, `Options`, `Compiler`, `HTMLEscape`,
  `URLEncode`.
- [Trim modes & literals](trim-modes.md) — the `trim_mode` string, the `<%% %%>`
  escapes, and magic comments.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at
[github.com/go-ruby-erb/erb](https://github.com/go-ruby-erb/erb).
