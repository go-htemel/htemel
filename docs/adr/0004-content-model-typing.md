# 0004 — Compile-time content models with wildcard escape hatches

**Status:** Accepted

## Context

With untyped children, `P(Div())` and `Ul(Div())` compiled and produced invalid HTML that
browsers silently repaired into a different tree. The spec data (ADR 0001) now carries
each element's categories and permitted children.

## Decision

Package `htemel` defines one marker interface per content category (`FlowContent`,
`PhrasingContent`, …; phrasing implies flow). Generated elements embed exported marker
structs for their categories, so user-defined component types can join a category by
embedding the same markers. Constructors, `Children`, `If` and `Text` take the element's
permitted children: a category interface, a generated per-element child interface for
mixed models (`UlChild`, `TableChild`, `SelectChild`), `TextContent` for text-only
elements, and nothing at all for empty elements. Transparent elements accept flow.
`template` is unchecked. SVG and MathML have closed per-package models; only their root
elements count as HTML embedded/phrasing/flow content.

`Group`, `Generic`, `Doctype`, `Each`, `ForEach`, `If`, `IfElse` and `Unchecked` return a
wildcard type that satisfies every content interface.

## Consequences

- Direct nesting mistakes are compile errors with readable messages
  (`*html.DivElement does not implement htemel.PhrasingContent`).
- **Known holes, accepted:** anything routed through `Each`/`Group`/`If` is unchecked,
  transparent elements accept flow (so `P(A(Div()))` compiles), descendant rules (no
  interactive content inside `button`, no nested `a`) are not expressed, and `template`
  is open. Dynamic content is the most common pattern, so the guarantee must be described
  as "catches direct nesting mistakes", not "validates HTML". ADR 0012 proposes a test-time
  validator to cover the rest.
- Cross-package wiring needs the generated `slots_*.go` files in package `htemel`, which
  is why `go generate` touches the root package.
