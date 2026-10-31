# 0010 — Move rare global attributes behind a `Globals` value

**Status:** Proposed — needs the owner's call

## Context

Each element has ~35 global attribute setters, ~20 built-ins and its own attributes:
around 90 methods on `Div()`. The wall is mostly globals nobody types daily
(`Itemprop`, `Itemref`, `Itemscope`, `Itemtype`, `Itemid`, `Enterkeyhint`,
`Writingsuggestions`, `Autocapitalize`, `Autocorrect`, `Spellcheck`, `Translate`,
`Inputmode`, `Inert`, `Nonce`, `Slot`, `Is`, `Part`, `Exportparts`, `Accesskey`,
`Draggable`, `Popover`, `Popovertarget*`, `Headingoffset`, `Headingreset`, …).

## Decision (proposed)

Keep the everyday globals as direct methods — `Id`, `Class`, `Style`, `Title`, `Lang`,
`Dir`, `Hidden`, `Tabindex`, `Role`, `Data`, `Aria`, `Attr*` — and move the rest into a
typed value applied in one call:

```go
html.Div().Globals(html.Globals{Itemprop: "name", Enterkeyhint: html.EnterkeyhintSend})
```

Struct fields carry the same enum types; zero values are omitted.

## Consequences

- Per-element methods drop from ~90 to ~45 without losing any attribute; the rare ones
  become *more* discoverable because they sit together in one struct.
- Breaking; migration is mechanical but touches generated docs and the `ui` library.
- Alternative considered and rejected: a sub-accessor (`Div().Global().Itemprop()`) breaks
  the chain's return type.
