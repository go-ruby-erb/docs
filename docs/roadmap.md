# Roadmap

`go-ruby-erb/erb` is grown **test-first**, each capability differential-tested
against MRI rather than built in isolation. The compiler — the deterministic,
interpreter-independent core of ERB — is **complete**; the rendering `eval` is
downstream by design.

| Stage | What | Status |
| --- | --- | --- |
| Tag scanner & literals | The MRI tag scanner for `<% %>` / `<%= %>` / `<%# %>` and the `<%%` / `%%>` literal escapes, split into the exact stags/etags `ERB::Compiler` produces. | **Done** |
| Trim modes | Every MRI `trim_mode` and its combinations — `-` explicit trim, `>` end-of-line, `<>` start-and-end, `%` percent-line (`%%` escaped). | **Done** |
| Compiler & text encoding | AST → Ruby source via MRI's put/insert/pre/post command wiring, with literal runs emitted through `String#dump` on the binary string for byte-exact round-tripping. | **Done** |
| Magic comments & `ERB::Util` | Leading `<%# coding: … %>` / `frozen_string_literal: …` (incl. the emacs `-*- … -*-` form) reflected in the emitted prefix, plus `HTMLEscape` / `URLEncode`. | **Done** |
| Differential oracle & coverage | A wide corpus compiled both here and by the system `ruby`, comparing emitted source **and** rendered output byte-for-byte; 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches. | **Done** |
| Rendering (`eval`) | `eval(compiled_src, binding)` against a live binding to produce the rendered string. **Downstream** — needs a Ruby interpreter; lives in the go-embedded-ruby consumer (rbgo), not this module. | Downstream |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter / no rendering.** The library emits the Ruby source and the
  magic comment; it never runs Ruby. The `eval` that produces the string is the
  consumer's job — see [Why a compiler, not a renderer](why.md).
- **Reference is MRI 4.0.5.** Byte-for-byte conformance targets that version's
  `ERB::Compiler` and `ERB::Util`; behaviour that changed across MRI releases is
  matched to this reference.
- **`String#dump` encoding fidelity** is to MRI's escaping of the **binary**
  string; callers needing a different encoding transcode at the boundary.
- **`ERB::Util` is `html_escape` + `url_encode`** (the `.h` / `.u` helpers); the
  full `ERB` object surface (`ERB.new`, `#result`, `#run`) is part of the
  rendering layer and therefore downstream.

See [Usage & API](api.md) for the surface and
[Trim modes & literals](trim-modes.md) for the scanning rules these stages build
out.
