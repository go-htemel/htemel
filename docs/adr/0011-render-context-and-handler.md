# 0011 — `RenderContext` and an HTTP handler helper

**Status:** Accepted (v0.4.0)

## Context

`Render(w)` takes no `context.Context`. Localisation, CSP nonces, request-scoped data,
cancellation and tracing must be threaded through component parameters by hand. ADR 0003
left a trap: direct `.Render` on an `http.ResponseWriter` is slow and can emit partial
output on error with a 200 status already sent.

## Decision (proposed)

- `Node` gains nothing; add `RenderContext(ctx, w, node)` that renders through a buffered
  writer and makes `ctx` available via a small `htemel.ContextOf(w)`/writer wrapper for
  components that need it (an explicit `ctx` parameter remains the recommended way to pass
  request data).
- `htemel.Handler(func(r *http.Request) Node)` and `htemel.Write(w, r, node)`: buffer,
  set status/content-type, write once; on error, log and send a 500 without a partial body.

## Consequences

- The obvious HTTP path becomes the safe path.
- Components stay plain functions; context is available but not implicit.

## Outcome

Implemented in v0.4.0 as `RenderContext(ctx, w, n)`, `ContextOf(w)`, `Write(w, r, n)` (buffers the whole page; on error writes nothing and returns the error, leaving the response to the caller) and `Handler(func(*http.Request) Node)` (200 on success; on error logs with the standard `log` package and sends a plain-text 500).
