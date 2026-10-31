# 0002 — Escaping by construction; raw text only where the spec says so

**Status:** Accepted

## Context

The first version escaped text but not attribute values (only `Data`/`Aria` were escaped),
so `Div().Title(`"><script>`)` injected markup. `Script().Text("a < b")` was escaped,
breaking JavaScript.

## Decision

Every string attribute value and every `Text` is escaped at render time. The only ways to
emit unescaped content are explicitly named: `TextUnsafe`, `AttrUnsafe`, `DataUnsafe`,
`AriaUnsafe`, and `Attribute{Unsafe: true}`. Elements the spec marks `raw_text`
(`script`, `style`) render `Text` unescaped and only guard a literal `</tag`. Fuzz tests
feed hostile strings through text and attribute setters and parse the output with the real
HTML parser to assert nothing escaped its context.

## Consequences

- Callers stop auditing strings; the unsafe paths are greppable by name.
- Escaped JSON in attributes renders as `&#34;`-heavy markup. It is correct HTML and
  Datastar reads it fine, but it looks noisy in view-source. Accepted.
- Untrusted URLs are still a caller concern (see ADR 0007).
