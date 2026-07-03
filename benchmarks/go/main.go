// SPDX-License-Identifier: BSD-3-Clause
package main

import (
	"fmt"
	"os"

	"github.com/go-ruby-erb/erb"
)

// Fixed, reproducible inputs shared byte-for-byte with benchmarks/ruby/erb.rb.
//
// template exercises every construct: <%= %> expression tags, <% %> control
// flow (each / if-else), a <%# %> comment, an <%%…%%> literal escape, and it is
// compiled under trim mode "-" (newline-trimming). The comparable Go op is
// erb.Compile (template -> Ruby source); MRI's equivalent is ERB.new(t).src.
const (
	template = `<!DOCTYPE html>
<html>
<head><title><%= @title %></title></head>
<body>
  <h1><%= ERB::Util.h(@heading) %></h1>
  <ul>
<% @items.each do |item| -%>
    <li id="item-<%= item[:id] %>"><%= ERB::Util.h(item[:name]) %> &mdash; <%= item[:qty] %></li>
<% end -%>
  </ul>
<%# footer section (not rendered) %>
<% if @items.empty? -%>
  <p>No items.</p>
<% else -%>
  <p>Total: <%= @items.size %> item(s)</p>
<% end -%>
  <footer>Use <%%= tag %%> to embed a literal.</footer>
</body>
</html>
`
	trimMode = "-"

	htmlIn = `a&b<c>"d"'e' café & naïve <script>alert(1)</script> 100% done`
	urlIn  = "a b/c?d=e#f&g=café 😀 ~-_.value"
)

// compile is the primary benchmarked op: template -> Ruby source.
func compile() string {
	src, _, err := erb.Compile(template, erb.Options{TrimMode: trimMode})
	if err != nil {
		panic(err)
	}
	return src
}

// verify prints each op's canonical output so it can be diffed byte-for-byte
// against the Ruby side before any timing is trusted.
func verify() {
	fmt.Print("=== compile ===\n")
	fmt.Print(compile())
	fmt.Print("=== html-escape ===\n")
	fmt.Print(erb.HTMLEscape(htmlIn))
	fmt.Print("\n=== url-encode ===\n")
	fmt.Print(erb.URLEncode(urlIn))
	fmt.Print("\n")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "verify" {
		verify()
		return
	}
	bench("compile", 2000, func() { sink = compile() })
	bench("html-escape", 5000, func() { sink = erb.HTMLEscape(htmlIn) })
	bench("url-encode", 5000, func() { sink = erb.URLEncode(urlIn) })
}
