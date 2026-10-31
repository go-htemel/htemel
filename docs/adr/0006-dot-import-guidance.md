# 0006 — Dot imports are a convenience, not the contract

**Status:** Accepted

## Context

The README shows `. "github.com/go-htemel/htemel/html"` for readability. Helpers added to
package `htemel` collided under dual dot-import (`Map` vs `<map>`), and `svg` shares names
with both `htemel` (`Text`) and `html` (`A`, `Title`, `Style`, `Script`, `Image`).

## Decision

New exported names in `htemel` are checked against every element name in every spec
before they are added (the collision that forced `Map` → `Each` is a test now). The
README recommends dot-importing at most `htemel` and `html` together and always aliasing
`svg`/`mathml`. The library never relies on dot imports in its own examples beyond that.

## Consequences

- Real code will mostly write `html.Div(...)`; that is fine and expected.
- **Sharp edge, accepted:** a future element name could still collide with an `htemel`
  helper; the check turns that into a generation-time failure rather than a user-visible
  one.
