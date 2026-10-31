# 0007 — URL attributes are escaped, not scheme-filtered

**Status:** Accepted

## Context

`A().Href("javascript:alert(1)")` renders as written. The spec data flags URL-valued
attributes, so filtering at the setter was possible.

## Decision

Setters keep taking `string` and only escape. `htemel.URL`, `MustURL` and `URLOr`
validate schemes (http, https, mailto, tel, sms, ftp, blob, `data:image/*` except
`svg+xml`, relative references; control characters rejected) for callers that hold
untrusted input. URL setters say so in their doc comment.

## Consequences

- No surprising rejections of legitimate app-specific schemes at render time.
- Responsibility sits with the caller; a `SafeURL` type on the setters remains an option
  if a real incident shows the helper is not being used.
