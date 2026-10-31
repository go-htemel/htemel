# 0013 — A helper for composite elements that wrap a control

**Status:** Proposed

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
