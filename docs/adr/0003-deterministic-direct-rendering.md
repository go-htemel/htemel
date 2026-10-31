# 0003 — Deterministic output, direct writes, buffered helpers

**Status:** Accepted

## Context

Attributes lived in a Go map, so output order changed between runs and a golden test
failed one time in thirty. Each element rendered its subtree into its own
`strings.Builder` and copied it into the parent: a page was copied once per level of
nesting.

## Decision

Attributes are an ordered list: first set fixes position, later sets update in place.
Elements write straight to the `io.Writer` they are given. `htemel.Render(w, node)` wraps
an unbuffered writer in a pooled buffer and flushes once; `RenderString`/`RenderBytes`
render into reused buffers. Generic attribute maps render in sorted key order.

## Consequences

- Output is byte-stable, which makes golden tests, ETags and caches reliable.
- A realistic page went from ~1,500 allocations to 9 and ~4× faster.
- **Sharp edge:** calling `node.Render(w)` directly on an unbuffered writer is slow
  (hundreds of syscalls) and, if a child returns an error mid-tree, leaves partial output
  in `w`. `htemel.Render` discards its buffer on error. The docs say so; ADR 0011 proposes
  an HTTP helper so the safe path is also the obvious one.
