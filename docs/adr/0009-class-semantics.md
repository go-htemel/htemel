# 0009 — Make `Class` additive

**Status:** Accepted (v0.4.0)

## Context

`Class(...)` replaces the class attribute; `AddClass(...)` appends. Every builder of the
`ui` library hit this once: `Class` on a component wiped its base classes. The `ui`
package grew `withClass` options purely to avoid it.

## Decision

Make `Class(...)` append (de-duplicating tokens) and add `SetClass(...)` for replacement.
`AddClass` becomes an alias for one release, then goes.

## Consequences

- The common intent ("also give it these classes") becomes the default.
- Breaking in behaviour, not signature: code that relied on replacement must switch to
  `SetClass`. A go vet-style check is not possible; document loudly in the CHANGELOG.

## Outcome

Implemented in v0.4.0: `Class` appends with token de-duplication and preserved order, `SetClass` replaces, `AddClass` is a deprecated alias, `ToggleClass` is unchanged. `Attr`/`Attrs`/datastar writes of `class` still replace.
