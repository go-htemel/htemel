# 0012 — Runtime content-model validator for tests

**Status:** Proposed

## Context

ADR 0004's static checks stop at wildcards, transparent elements and descendant rules.
Most dynamic content goes through `Each`, so the most common composition is unchecked.

## Decision (proposed)

Add `htemel/check` (or a test helper) that walks a rendered tree against the spec's
content models — including transparent inheritance and the descendant rules the type
system cannot express — and reports violations with a path. Off in production; used in
tests and the showcase's page tests.

## Consequences

- Turns the known holes into a diagnostic without changing the public API.
- Needs the generated packages to expose their content-model data (a small generated table).
