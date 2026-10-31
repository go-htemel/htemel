# 0001 — Generate the element API from standards data

**Status:** Accepted (2026-10-31)

## Context

The original `html` package was generated from a hand-written attribute list in
`go-htemel/spec`. It carried errors (`td span`, a misspelled `shadowrootdelegatesfocus`,
`download` as a boolean, `textarea rows` as a string), omitted `sup`, could not tell URL
attributes from strings, and had no content-model information. Regenerating depended on
scraping the live WHATWG page.

## Decision

`go-htemel/spec` builds its documents offline with `specgen2` from three committed inputs:
`@webref/elements` + `@webref/idl` (which elements exist; attribute names and reflected
types), `@mdn/browser-compat-data` (deprecation, extra attributes), and a pinned extract of
the WHATWG HTML index tables (attribute value types, content categories, permitted
children). A small overlay file holds judgment calls, each with a written reason. Every
attribute carries provenance. htemel consumes the resulting JSON and never hand-edits
generated files; CI fails if the generated tree differs from the templates and data.

SVG and MathML elements come from the same pipeline (webref lists them), so they cost
one extra `go:generate` line each.

## Consequences

- Correctness is traceable: a wrong setter is a data bug with a known source, not a typo.
- The standard moving is a command (`fetch`, `generate`, `report`) and a reviewable diff.
- The WHATWG index parser is matched to the current page markup. The snapshot is pinned,
  so nothing breaks silently, but re-pinning may need parser work. Someone owns that.
- Every htemel release is coupled to a spec release. Keep both tags moving together.
- Where the data cannot express something (raw-text elements, `template`'s content model,
  `download` being bool-or-string) the overlay decides. Those entries are the places to
  look first when behaviour surprises.
