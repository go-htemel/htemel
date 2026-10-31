# 0010 — Move rare global attributes behind a `Globals` value

**Status:** Accepted (v0.4.0)

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

## Outcome

Implemented in v0.4.0 for the `html` package only (SVG/MathML "globals" are presentation attributes typed constantly, so they stay methods). Final keep-list (`generators.HTMLDirectGlobals`): `id`, `class`, `style`, `title`, `lang`, `dir`, `hidden`, `tabindex`, `role`, `autofocus`, `draggable`, `popover`, `nonce`, `data`, `aria`. The other 21 globals (`accesskey`, `autocapitalize`, `autocorrect`, `contenteditable`, `enterkeyhint`, `exportparts`, `headingoffset`, `headingreset`, `inert`, `inputmode`, `is`, `itemid`, `itemprop`, `itemref`, `itemscope`, `itemtype`, `part`, `slot`, `spellcheck`, `translate`, `writingsuggestions`) are fields of `html.Globals` (shared enum types, `string`, `bool`, `int`, `[]string` for token lists, `string` for the one-character `accesskey`), applied in declaration order with zero values skipped. `*html.DivElement` went from 54 to 32 methods.
