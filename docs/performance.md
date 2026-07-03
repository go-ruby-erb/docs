# Performance

`go-ruby-erb/erb` is the pure-Go, MRI-compatible **ERB template compiler** that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `erb`. This
page records a **comparative, library-level benchmark** of that module against
the reference Ruby runtimes, part of the ecosystem-wide per-module parity suite.

## What "compile" maps to

The library is a **compiler, not a renderer**: `erb.Compile` turns an ERB
template into the **Ruby source** that renders it, and the eval of that source
against a binding stays in `rbgo` (exactly as MRI splits `ERB.new(...).src` from
`.result`). The exactly-comparable Ruby op is therefore
**`ERB.new(template, trim_mode:).src`** — MRI's own template->source step. Two
`ERB::Util` helpers the library implements, `html_escape` and `url_encode`, are
measured directly against `ERB::Util.html_escape` / `ERB::Util.url_encode`.

**Rendering is not benchmarked**: the Go library does not execute, so timing
`ERB#result` would measure an interpreter's eval loop, not this library —
apples-to-oranges. Only ops the pure-Go library actually performs are timed.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This measures the **pure-Go library directly, through its Go API**, isolated
from any interpreter dispatch, answering the parity question head-on: *is the
pure-Go compiler as fast as the reference runtime's own `ERB`?* The **same
workload, same fixed inputs, same iteration counts** run through the Go library
and through each reference runtime's stdlib; outputs were checked
**byte-identical to MRI** before any timing (see *Reproduce* below).

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
  All runtimes measured **on the host**, no VM.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` (the oracle) · MRI + YJIT ·
  JRuby 10.1.0.0 (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Fixed inputs (reproducible, never variable):** one representative HTML
  template exercising `<%= %>` expression tags, `<% %>` control flow
  (`each` + `if`/`else`), a `<%# %>` comment and an `<%%…%%>` literal, compiled
  under trim mode `-`; one HTML string (`a&b<c>"d"'e' café & naïve <script>…`)
  for `html_escape`; one URL string (`a b/c?d=e#f&g=café 😀 ~-_.value`) for
  `url_encode`.
- **Operations:** `compile` — `erb.Compile` vs `ERB.new(t, trim_mode: "-").src`;
  `html-escape` — `erb.HTMLEscape` vs `ERB::Util.html_escape`; `url-encode` —
  `erb.URLEncode` vs `ERB::Util.url_encode`.
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.

#### compile

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 6124.7 | 0.15× |
| MRI | 40915.5 | 1.00× |
| MRI + YJIT | 30776.5 | 0.75× |
| JRuby | 33343.7 | 0.81× |
| TruffleRuby | 13190.5 | 0.32× |

#### html-escape

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 133.1 | 0.83× |
| MRI | 160.0 | 1.00× |
| MRI + YJIT | 111.4 | 0.70× |
| JRuby | 165.6 | 1.03× |
| TruffleRuby | 1239.6 | 7.75× |

#### url-encode

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 130.3 | 0.47× |
| MRI | 280.0 | 1.00× |
| MRI + YJIT | 229.4 | 0.82× |
| JRuby | 274.5 | 0.98× |
| TruffleRuby | 13043.9 | 46.59× |

**Compile** is where the pure-Go library pulls decisively ahead: turning the
template into Ruby source is **~6.7× faster than MRI** (0.15×) and ~5× faster
than MRI + YJIT, because MRI's `ERB::Compiler` builds the source through a chain
of Ruby `Scanner`/`Buffer` objects while the Go port scans the template with a
single tight state machine. YJIT (0.75×) and JRuby (0.81×) close some of that
gap but stay behind; TruffleRuby's compile (0.32×) is its strongest row. The
`ERB::Util` helpers are much closer to parity — pure byte-walking work where
MRI's C is already tight: `html_escape` is a shade faster than MRI (0.83×) and
`url_encode` ~2× faster (0.47×), both in YJIT/JRuby territory. TruffleRuby's two
util rows are **cold-JIT** figures (7.75× / 46.59×): these sub-microsecond ops
never reach the pass count Graal needs to compile them, so they read as
interpreted, not steady-state — see the warning below.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-erb/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via `go.mod`
    pseudo-version — no `replace`), the equivalent `ruby/erb.rb` workload, and
    `run.sh`. Run `bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the pass
    budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries. Both
    drivers accept a `verify` argument
    (`(cd benchmarks/go && GOWORK=off go run . verify)` vs
    `ruby benchmarks/ruby/erb.rb verify`) that prints each operation's canonical
    output — the compiled source, the escaped string and the encoded string were
    confirmed **byte-identical** before timing.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput — most visibly TruffleRuby on the sub-microsecond `html-escape` /
    `url-encode` rows, which are cold-JIT figures, not steady-state. Those
    sub-microsecond rows also carry the most relative noise; treat their ratios
    as order-of-magnitude. Every number here is a **real measured value** from
    the dated run above — nothing is fabricated, estimated, or cherry-picked. The
    go-ruby column is the pure-Go library; every other column is that
    interpreter's own `ERB` / `ERB::Util` doing the equivalent work. Rendering is
    absent by design: the library compiles, it does not execute.
