# Why a compiler, not a renderer

ERB has two halves. The first is **deterministic**: given a template string and a
trim mode, scanning the tags and emitting the Ruby source that would render it is
a pure function of the input — no variables, no bindings, no interpreter. The
second is the **`eval`**: running that emitted Ruby source against a binding to
produce the final string. That second half needs a Ruby interpreter; the first
does not.

`go-ruby-erb/erb` is exactly the first half. It reimplements MRI's
`ERB::Compiler` in pure Go, so the deterministic part of ERB runs as a
dependency-free, CGO-free Go library, and the interpreter-bound part stays where
the interpreter lives.

## What MRI's `ERB::Compiler` does

Given a template like:

```erb
Hello <%= name %>!
```

MRI's compiler does **not** render it. It produces the Ruby program that, when
`eval`'d against a binding, renders it:

```ruby
#coding:UTF-8
_erbout = +''; _erbout.<< "Hello ".freeze; _erbout.<<(( name ).to_s); _erbout.<< "!\n".freeze
; _erbout
```

Everything in that transformation is mechanical and interpreter-independent:

- **tag scanning** — splitting the source into literal runs and `<% %>` /
  `<%= %>` / `<%# %>` tags, including the `<%%` / `%%>` literal escapes;
- **trim modes** — applying the `trim_mode` string (`-`, `>`, `<>`, `%`) to
  decide which newlines around tags survive;
- **text encoding** — emitting each literal run through Ruby's `String#dump` on
  the **binary** string so quotes, newlines, control bytes and multi-byte UTF-8
  round-trip byte-for-byte (`héllo` → `"h\xC3\xA9llo"`);
- **magic comments** — detecting a leading `<%# coding: … %>` /
  `frozen_string_literal: …` (including the emacs `-*- … -*-` form) and
  reflecting it in the emitted `#coding:` / `#frozen-string-literal:` prefix.

None of that needs to *run* Ruby. It only needs to **be** Ruby's algorithm,
faithfully — which is what this library is.

## Where the interpreter is still required

The final step:

```ruby
eval(compiled_src, binding)   # -> "Hello World!\n"
```

needs the names in the template (`name`, here) resolved against a live binding,
and arbitrary Ruby expressions evaluated. That is the interpreter's job, and it
stays in the consumer — for example [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)'s
`rbgo`, which hands this library a template, gets back the compiled source, and
`eval`s it in its own VM.

> **This library compiles; the host evaluates.** The split is what lets the
> deterministic 90% of ERB ship as a pure-Go, statically-linkable module while
> the interpreter-bound 10% lives with the interpreter.

## Why pure Go matters here

Because the compiler is CGO-free and dependency-free, it:

- cross-compiles to every Go target with no C toolchain, and links into a single
  static binary;
- has **no dependency on the Ruby runtime** — the dependency runs the other way;
- can be differentially tested against the `ruby` binary wherever one is on
  `PATH`, while the cross-arch lanes (where `ruby` is absent) still validate the
  compiler itself.

See [Usage & API](api.md) for the surface, and [Trim modes & literals](trim-modes.md)
for the scanning rules this faithfulness rests on.
