# Trim modes & literals

ERB's whitespace handling and its literal escapes are where most of the
byte-for-byte subtlety lives. `go-ruby-erb/erb` follows MRI's `ERB::Compiler`
scanner exactly, so the emitted source — and therefore the rendered output —
matches the reference.

## Tag kinds

| Tag | Meaning | Emitted as |
| --- | --- | --- |
| `<% code %>` | run Ruby, emit nothing | the bare statement |
| `<%= expr %>` | evaluate and insert | `_erbout.<<(( expr ).to_s)` |
| `<%# comment %>` | comment, discarded | nothing |

## Literals: `<%%` and `%%>`

A doubled opening tag is a **literal** `<%`, and inside a tag a doubled `%>` is a
literal `%>`. The scanner unescapes them the way MRI does:

| In the template | Renders as |
| --- | --- |
| `<%%` | `<%` |
| `%%>` | `%>` |

So `Rate: <%%= n %>` renders the literal text `Rate: <%= n %>` rather than
treating it as a tag.

## Trim modes

The `TrimMode` option is MRI's `trim_mode` string. Each character turns on one
rule; they **combine** (e.g. `"<>-"`).

| Mode | Rule |
| --- | --- |
| `-` | **Explicit trim.** `-%>` strips the immediately-following newline; `<%-` strips leading whitespace before the tag. The trim is opt-in per tag. |
| `>` | Strips the newline after a tag that **ends its line**. |
| `<>` | Strips the newline only when the tag **both starts and ends** the line. |
| `%` | A line whose first non-blank char is `%` is a **code line** (`% code` ≡ `<% code %>`); `%%` at line start is an escaped literal `%`. |

!!! note "Combinations"
    Modes compose. `"-"` is the most common: it leaves normal newlines alone and
    only trims where you write `-%>`, which is what most templates want. `"%>"`
    or `"%<>"` combine the percent-line syntax with an automatic end-of-line
    trim.

### Example

```go
src, _, _ := erb.Compile(
	"<ul>\n<% items.each do |i| -%>\n  <li><%= i %></li>\n<% end -%>\n</ul>\n",
	erb.Options{TrimMode: "-"})
```

The `-%>` on the `each` and `end` lines strips the newline that would otherwise
follow each control tag, so the rendered list has no blank lines between items —
byte-for-byte what MRI produces for the same template and trim mode.

## Binary-exact text encoding

Each literal run between tags is emitted via Ruby's `String#dump` on the
**binary** string, not via naive quoting. That is what makes embedded quotes,
newlines, control bytes, and multi-byte UTF-8 round-trip exactly:

| Literal run | Emitted |
| --- | --- |
| `héllo` | `"h\xC3\xA9llo"` |
| a line with `"` and a tab | the `String#dump` escaping MRI uses |

Because the encoding is byte-level and matches `String#dump`, the compiled source
is identical to MRI's down to the escape sequences — which is exactly what the
[differential oracle](api.md#mri-conformance) checks.

## Magic comments

A leading `<%# coding: … %>` or `<%# frozen_string_literal: … %>` — including the
emacs `-*- coding: … -*-` form — is detected and reflected in the emitted
prefix:

| In the template | Emitted prefix |
| --- | --- |
| (none) | `#coding:UTF-8` |
| `<%# coding: us-ascii %>` | `#coding:us-ascii` |
| `<%# frozen_string_literal: true %>` | `#frozen-string-literal:true` |

The `magicComment` return value of [`Compile`](api.md#compile) reports this line
on its own, and `src` already carries it as its first line.
