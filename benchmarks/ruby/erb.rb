# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
require "erb"
require_relative "_harness"

$stdout.binmode # keep captured output LF-clean on every platform

# Fixed, reproducible inputs — byte-for-byte identical to benchmarks/go/main.go.
# The template exercises <%= %> expression tags, <% %> control flow (each /
# if-else), a <%# %> comment, and an <%%…%%> literal escape, compiled under
# trim mode "-". The comparable op is ERB.new(t).src (template -> Ruby source),
# matching the Go library's erb.Compile; rendering/eval is out of scope (the
# Go library does not execute — that stays in rbgo).
TEMPLATE = <<'TPL'
<!DOCTYPE html>
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
TPL

TRIM_MODE = "-"
HTML_IN   = %(a&b<c>"d"'e' café & naïve <script>alert(1)</script> 100% done)
URL_IN    = "a b/c?d=e#f&g=café \u{1F600} ~-_.value"

# compile is the primary benchmarked op: template -> Ruby source (the .src).
def compile
  ERB.new(TEMPLATE, trim_mode: TRIM_MODE).src
end

if ARGV[0] == "verify"
  print "=== compile ===\n"
  print compile
  print "=== html-escape ===\n"
  print ERB::Util.html_escape(HTML_IN)
  print "\n=== url-encode ===\n"
  print ERB::Util.url_encode(URL_IN)
  print "\n"
  exit
end

bench("compile",     2000) { compile }
bench("html-escape", 5000) { ERB::Util.html_escape(HTML_IN) }
bench("url-encode",  5000) { ERB::Util.url_encode(URL_IN) }
