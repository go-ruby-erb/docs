<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-erb` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-erb/erb`
library** against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby,
TruffleRuby). It measures the **library primitive** through its Go API, isolated
from any interpreter, so the numbers answer: *is the pure-Go ERB compiler as
fast as the reference runtime's own `ERB`?*

## What "compile" means here

`go-ruby-erb/erb` is a **template-compiler**, not a renderer: `erb.Compile`
turns an ERB template into the **Ruby source** that would render it (the eval of
that source against a binding stays in [rbgo](https://github.com/go-embedded-ruby/ruby)).
The exactly-comparable Ruby op is therefore **`ERB.new(template, trim_mode:).src`**
— MRI's own template→source step — *not* `.result`. Rendering is deliberately
**not** benchmarked: the Go library does not execute, so timing `.result` would
measure an interpreter's eval loop, not this library. Two `ERB::Util` helpers
that the library *does* implement — `html_escape` and `url_encode` — are
measured directly against `ERB::Util.html_escape` / `ERB::Util.url_encode`.

## Layout

- `go/`          — self-contained Go driver; `go.mod` pins the published library
  by pseudo-version (no `replace`). Any built `go/bench` binary is git-ignored.
- `ruby/erb.rb`  — the equivalent workload; `ruby/_harness.rb` is the shared timer.
- `run.sh`       — runs every available runtime and prints one Markdown table per
  sub-benchmark (ns/op + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Verify (outputs identical to MRI)

Both drivers accept a `verify` argument that prints each operation's canonical
output for the **fixed** template and escape inputs, so correctness is checked
before any timing is trusted:

```sh
diff <(cd go && GOWORK=off go run . verify) <(ruby ruby/erb.rb verify)   # must be empty
```

The compiled source, the HTML-escaped string and the URL-encoded string were all
confirmed **byte-identical** between the Go library and MRI before timing.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build **identical fixed inputs**
(one representative template with `<%= %>` / `<% %>` / `<%# %>` / `<%%…%%>` under
trim mode `-`, plus one HTML string and one URL string). Results are published,
dated, in `../docs/performance.md`.
