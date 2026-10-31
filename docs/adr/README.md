# Architecture decision records

One file per decision, numbered in the order they were made. Status is one of
**Accepted**, **Proposed** (agreed direction, not implemented), **Superseded** or
**Rejected**. Each record states the context, the decision, and the consequences we
know about — including the sharp edges a decision created, so that "accept" is a
choice rather than an accident.

| # | Title | Status |
|---|---|---|
| [0001](0001-generate-elements-from-spec-data.md) | Generate the element API from standards data | Accepted |
| [0002](0002-escaping-by-construction.md) | Escaping by construction; raw text only where the spec says so | Accepted |
| [0003](0003-deterministic-direct-rendering.md) | Deterministic output, direct writes, buffered helpers | Accepted |
| [0004](0004-content-model-typing.md) | Compile-time content models with wildcard escape hatches | Accepted |
| [0005](0005-shared-enum-types.md) | One enum type per attribute, variadic setters for empty-allowed enums | Accepted |
| [0006](0006-dot-import-guidance.md) | Dot imports are a convenience, not the contract | Accepted |
| [0007](0007-no-url-scheme-filtering.md) | URL attributes are escaped, not scheme-filtered | Accepted |
| [0008](0008-retire-per-element-conditionals.md) | Retire `XxxIf`/`XxxTernary` and the `If*` methods | Proposed |
| [0009](0009-class-semantics.md) | Make `Class` additive | Proposed |
| [0010](0010-rare-globals-behind-a-struct.md) | Move rare global attributes behind a `Globals` value | Proposed |
| [0011](0011-render-context-and-handler.md) | `RenderContext` and an HTTP handler helper | Proposed |
| [0012](0012-debug-content-model-validator.md) | Runtime content-model validator for tests | Proposed |
| [0013](0013-delegating-element-helper.md) | A helper for composite elements that wrap a control | Proposed |
