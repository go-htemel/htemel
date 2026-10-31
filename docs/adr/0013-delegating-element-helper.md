# 0013 — A helper for composite elements that wrap a control

**Status:** Rejected

## Context

Composite components (`ui.Select`, `InputGroup`, `TagInput`, `FileUpload`,
`MarkdownEditor`) wrap a real control in presentational markup and must forward `Attrs`,
`SetAttr`, `SetAttrUnsafe` (and `Field`'s id/aria wiring) to the inner element. Each one
re-implements the same four methods.

## Decision (proposed)

Provide in `htemel` a generic `Delegating[T Element]` embedding helper (or a generated
`WrapperOf*` type) that implements `Element`, the category markers of the wrapper, and
forwards attribute methods to a `Control()`; `ui.CompositeControl` becomes a thin
interface over it.

## Consequences

- Composites shrink to "build the markup, name the control".
- Generics on method return types (`Attrs(...) *Self`) need care; worst case the helper
  returns the interface and composites keep one typed `Attrs` each.

## Outcome

Prototyped for v0.4.0 and not shipped. Go cannot embed a type parameter, so a generic wrapper cannot promote the outer element's markers. A `Composite{Outer Node; Inner Element}` helper cannot sit beside an embedded `*html.DivElement` (the promoted `Render`/`SetAttr` are ambiguous, so the type no longer implements `htemel.Element`). Used instead of the outer element, it drops the wrapper's builder methods and still needs a typed `Attrs` and `Control` per composite. It saves about two lines per composite, below the bar, so the README documents the three-method pattern instead.
