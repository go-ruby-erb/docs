# Usage & API

The public API lives at the module root (`github.com/go-ruby-erb/erb`). It is
**Ruby-shaped but Go-idiomatic**: `Compile` mirrors MRI's
`ERB::Compiler#compile` two-value contract (source + magic comment), but the
surface follows Go conventions — an explicit `error`, value types, no global
state.

!!! success "Status: implemented"
    The compiler is built and importable as `github.com/go-ruby-erb/erb`. The
    final `eval(compiled_src, binding)` that renders the string lives downstream
    in the consumer (e.g. rbgo); see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-erb/erb
```

## Compiling a template

```go
package main

import (
	"fmt"

	"github.com/go-ruby-erb/erb"
)

func main() {
	src, magic, err := erb.Compile("Hello <%= name %>!\n", erb.Options{})
	if err != nil {
		panic(err)
	}
	fmt.Print(magic) // #coding:UTF-8
	fmt.Println(src)
	// #coding:UTF-8
	// _erbout = +''; _erbout.<< "Hello ".freeze; _erbout.<<(( name ).to_s); _erbout.<< "!\n".freeze
	// ; _erbout
	//
	// src already carries the #coding prefix; hand it to a Ruby interpreter:
	//   eval(src, binding)  ->  "Hello World!\n"
}
```

With a trim mode and a custom buffer variable:

```go
src, _, _ := erb.Compile(
	"<ul>\n<% items.each do |i| -%>\n  <li><%= i %></li>\n<% end -%>\n</ul>\n",
	erb.Options{TrimMode: "-", EOutVar: "buf"})
```

## `ERB::Util` helpers

```go
fmt.Println(erb.HTMLEscape(`<a href="x">it's</a>`))
// &lt;a href=&quot;x&quot;&gt;it&#39;s&lt;/a&gt;
fmt.Println(erb.URLEncode("100% & more"))
// 100%25%20%26%20more
```

## Shape

### `Compile`

```go
// Compile returns the Ruby source that, when eval'd against a binding, renders
// the template, plus the magic-encoding comment line — matching MRI's
// ERB::Compiler#compile two-value contract.
func Compile(template string, opts Options) (src, magicComment string, err error)
```

- **`src`** — the emitted Ruby program (it already carries the `#coding:` prefix).
- **`magicComment`** — the magic-encoding comment line on its own (`#coding:UTF-8`
  by default, or whatever a leading `<%# coding: … %>` / `frozen_string_literal:`
  comment requests).
- **`err`** — non-nil for a malformed template (e.g. an unterminated tag).

### `Options`

```go
type Options struct {
	TrimMode string // MRI trim_mode: "", "-", ">", "<>", "%" and combinations
	EOutVar  string // output buffer var name; default "_erbout"
}
```

- **`TrimMode`** — the MRI `trim_mode` string; see
  [Trim modes & literals](trim-modes.md). The empty string is "no trimming".
- **`EOutVar`** — the name of the output-buffer variable in the emitted source
  (Ruby's `ERB.new(..., eoutvar:)`); defaults to `_erbout`.

### `ERB::Util`

```go
func HTMLEscape(s string) string // ERB::Util.html_escape / .h
func URLEncode(s string) string  // ERB::Util.url_encode / .u
```

- **`HTMLEscape`** — replaces `& < > " '` with their entities (`'` → `&#39;`),
  matching `ERB::Util.html_escape`.
- **`URLEncode`** — percent-encodes everything but `A-Za-z0-9-_.~`, with
  **upper-case** hex, matching `ERB::Util.url_encode`.

### `Compiler` (low-level)

For hosts that need MRI's put/insert/pre/post command wiring directly (most
callers use `Compile`):

```go
type Compiler struct {
	PutCmd, InsertCmd string   // the << / <<(( … )).to_s commands
	PreCmd, PostCmd   []string // statements emitted before / after the body
	EOutVar           string   // output buffer var name
}

func NewCompiler(trimMode string) *Compiler
func (c *Compiler) Compile(s string) (src, magicComment string, err error)
```

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** compiles a
wide template corpus — every tag kind, the `<%% %%>` literals, all trim modes,
multiline / quoted / unicode bodies — both with this library and with the system
`ruby`, comparing the **emitted Ruby source** *and* the **rendered result**
byte-for-byte, plus `ERB::Util` against MRI. The reference is **MRI 4.0.5**. The
oracle tests skip themselves where `ruby` is not on `PATH` (e.g. the qemu arch
lanes), so the cross-arch builds still validate the compiler.

## Relationship to Ruby

The consumer hands a template to `Compile`, gets back the Ruby source, and
`eval`s it against a binding in its own VM —
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)'s `rbgo` is the
reference consumer. Because `Compile` already produces MRI-byte-exact source with
the right magic comment, wiring it into a host is a mechanical step, not a
reimplementation: the library does the compiling, the host does the evaluating.
