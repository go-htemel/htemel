# 0005 — One enum type per attribute, variadic setters for empty-allowed enums

**Status:** Accepted

## Context

Each element had its own enum types (`DivPopover`, `SpanPopover`, … ×112): 1,331 types in
`html`, each with its own constants. Enum attributes that allow the empty value
(`hidden`, `contenteditable`) needed `Hidden(DivHiddenEmpty)` to render the bare attribute.

## Decision

When every element shares an attribute's allowed set, generate one type per package
(`html.Popover`, `PopoverManual`) in `enums.go`; keep per-element types only where the sets
differ (`type` on input/button/script; a few SVG filter attributes). Enum setters whose
attribute allows the empty value are variadic: `Hidden()` renders the bare attribute,
`Hidden(HiddenUntilFound)` the value, last argument wins.

## Consequences

- `html` enum types: 1,331 → 40. Autocomplete and docs are far less noisy.
- A few names are still awkward (`InputTypeDatetimeLocal`, `Accesskey(rune)` for a
  one-character attribute). Accepted for now; revisit with ADR 0010.
