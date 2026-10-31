# 0008 — Retire `XxxIf`/`XxxTernary` and the `If*` methods

**Status:** Accepted (v0.4.0)

## Context

Every element has `DivIf(cond, children...)`, `DivTernary(cond, a, b)`, and methods
`If`, `IfElse`, `IfThen`: 224 package-level functions plus 336 methods that predate the
package-level `htemel.If`/`IfElse` and duplicate them with worse discoverability.
They were the single largest avoidable contributor to the `html` API surface.

## Decision

Remove them in the next breaking release. `htemel.If(cond, html.Div(...))`,
`htemel.IfElse`, and `.With(func(*DivElement))` cover every use. Keep `.If(cond, child)`
on elements only if a survey of `ui` shows it is used widely; otherwise remove it too.

## Consequences

- ~560 fewer signatures; `Div().` autocomplete loses three entries, the package index
  loses two per element.
- Breaking; mechanical to migrate (`ui` has a handful of call sites).

## Outcome

Implemented in v0.4.0. The `ui` survey found 0 call sites of the element methods `.If`, `.IfElse` and `.IfThen` (and none of `XxxIf`/`XxxTernary`), so `.If` was removed as well; `htemel.If`/`IfElse`/`Each`/`ForEach` and `.With` remain.
