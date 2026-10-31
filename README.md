[![Go Reference](https://pkg.go.dev/badge/github.com/go-htemel/htemel.svg)](https://pkg.go.dev/github.com/go-htemel/htemel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Version](https://img.shields.io/github/v/release/go-htemel/htemel?sort=semver)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-htemel/htemel)](https://goreportcard.com/report/github.com/go-htemel/htemel)

# htemel

Pronounced like "H-temel", **htemel** is a Go library for generating web documents using pure Go functions, avoiding traditional templating languages. The library is designed for extensibility, making it easy to add custom functionality or support for bespoke web components.

## Installation

To install htemel, run the following commands:

```shell
go get github.com/go-htemel/htemel
```

## Example

Using a `.` (dot) import for namespaced element packages improves readability but is optional.

### Basic Usage

```go
package main

import (
    "os"

    . "github.com/go-htemel/htemel"
    . "github.com/go-htemel/htemel/html"
)

func MainLayout(children ...Node) Node {
    return Group(
        Doctype(),
        Html().Class("h-dvh bg-gray-200").Lang("en").Children(
            Head(
                Meta().Charset("UTF-8"),
                Meta().Name("viewport").Content("width=device-width, initial-scale=1.0"),
                Title().Text("Example htemel Page"),
                Link().Href("site.css").Rel("stylesheet"),
            ),
            Body().Id("app-content").Children(children...),
        ),
    )
}

func Navigation(menuItems ...Node) Node {
    return Nav(
        // Child elements can be added using the Children method
        Ul().Class("flex list-none").Children(
            Group(menuItems...),
        ),
    ).Id("main-navigation")
}

func main() {
    loggedIn := true
    // Or, if you prefer, children can be passed directly to the base element
    page := MainLayout(
        DivTernary(
            loggedIn,
            Group(
                Navigation(),
                Div(
                    P().Text("Welcome back!"),
                ),
            ),
            Div(
                P().Text("Please log in."),
            ),
        ),
    )

    // htemel.Render buffers writes; use page.Render(w) directly only when w is already buffered.
    Render(os.Stdout, page)
}
```

**Output**:

```html
<!DOCTYPE html><html class="h-dvh bg-gray-200" lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Example htemel Page</title><link href="site.css" rel="stylesheet"></head><body id="app-content"><div><nav id="main-navigation"><ul class="flex list-none"></ul></nav><div><p>Welcome back!</p></div></div></body></html>
```

The renderer does not add indentation or newlines.

### Rendering

Every `Node` has `Render(w io.Writer) error`, which writes directly to `w` with no buffering of its own. When `w` is an unbuffered destination such as an `http.ResponseWriter` or `os.Stdout`, use `htemel.Render(w, node)`, which wraps it in a buffer and flushes once, or `htemel.RenderString` / `htemel.RenderBytes` to get the output in memory. Calling `node.Render(w)` directly is fine when `w` is already a `*bytes.Buffer`, `*strings.Builder` or `*bufio.Writer`.

### Controlling Element Placement

Chained calls respect the order of elements in the document, for example:

```go
package main

import (
    "os"

    . "github.com/go-htemel/htemel/html"
)

func main() {
    div := Div().
        Children(
            P().Text("Start Text").Children(
                Span().Text("Spanned Text"),
            ).Text("End Text"),
        )

    div.Render(os.Stdout)
}
```

**Output**:

```html
<div><p>Start Text<span>Spanned Text</span>End Text</p></div>
```

### Writing Components

A component is a plain Go function that returns a `Node`. Arguments are ordinary Go values, and children are passed as `...Node`:

```go
func Card(title string, children ...Node) Node {
    return Div().Class("card").Children(
        H2().Text(title),
        Div().Class("card-body").Children(children...),
    )
}
```

- Use `Each` to build repeated content. It renders a function for each item at render time, so no intermediate slice of nodes is built:

  ```go
  func ItemList(items []string) Node {
      return Ul().Class("list").Children(
          Each(items, func(item string) Node {
              return Li().Text(item)
          }),
      )
  }
  ```

  `EachIndex` also passes the index, `ForEach` and `ForEach2` range over an `iter.Seq` or `iter.Seq2` such as `slices.Values` or `maps.All`, and `If` and `IfElse` choose between nodes without a separate `if` block.

- Use `Group` to return several sibling nodes without a wrapping element.

`With` is still available when you need the typed element itself inside a function. The [webapp example](examples/webapp/views/layout.go) shows these patterns in a complete application: a shared layout, a navigation component built with `Each`, an inline SVG icon, and a page wrapper that accepts children.

#### Forwarding Attributes

Every generated element has an `Attrs` method that takes `htemel.Attribute` values, so a component can accept attributes from its caller and pass them on. `Attr` and `AttrUnsafe` build an attribute, and `If` returns the attribute only when the condition holds:

```go
func Button(label string, attrs ...htemel.Attribute) htemel.Node {
    return html.Button().Attrs(attrs...).Text(label)
}

// At the call site:
htemel.RenderString(Button("Save",
    htemel.Attr("data-x", "y"),
    htemel.Attr("aria-label", `say "hi" & <bye>`),
    htemel.Attr("title", "shown").If(false),
    htemel.Attr("hidden", "").If(true),
))
```

**Output**:

```html
<button data-x="y" aria-label="say &#34;hi&#34; &amp; &lt;bye&gt;" hidden>Save</button>
```

An empty value renders a bare attribute, and the zero attribute that `If(false)` returns is skipped.

## SVG

The `svg` package generates the SVG elements, with tag and attribute names kept in their SVG case (`viewBox`, `linearGradient`, `feGaussianBlur`). SVG nodes are plain `Node` values, so they nest directly inside HTML elements.

Import `svg` under its package name rather than with a `.` import. It shares names with both `htemel` (`Text`) and `html` (`A`, `Title`, `Style`, `Script`, `Image`), so a dot import would conflict.

```go
func PlusIcon() Node {
    return svg.Svg().
        ViewBox("0 0 24 24").
        Attr("width", "24").Attr("height", "24").
        Children(
            svg.Path().
                D("M4 12h16M12 4v16").
                Fill("none").
                Stroke("currentColor").
                StrokeWidth("2").
                StrokeLinecap(svg.PathStrokeLinecapRound),
        )
}

// Inside an HTML element:
Button().Children(PlusIcon(), Text("Add"))
```

**Output**:

```html
<button><svg viewBox="0 0 24 24" width="24" height="24"><path d="M4 12h16M12 4v16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path></svg>Add</button>
```

- Presentation attributes such as `fill`, `stroke` and `opacity` are available on every element. Animation timing uses `fill="freeze"`, which collides with the paint `fill` method, so set it with `.Attr("fill", "freeze")`.
- Reference other elements with `svg.Use().Href("#id")`, and define them in `svg.Defs()`.
- The element and attribute list is hand-curated in `specs/svg.json` and generated into `svg/`.

## Datastar

The `datastar` package provides typed helpers for [Datastar](https://data-star.dev) attributes and action expressions. Each helper returns an `htemel.Attribute`, which any generated element accepts through `Attrs`:

```go
html.Div().Attrs(datastar.Signals(map[string]any{"count": 0})).Children(
    html.Button().Attrs(
        datastar.OnClick(datastar.Post("/inc"), datastar.Debounce(300*time.Millisecond)),
        datastar.Indicator("busy"),
        datastar.Attr("disabled", "$busy"),
    ).Text("+"),
    html.Span().Attrs(datastar.Text("$count")),
)
```

**Output**:

```html
<div data-signals="{&#34;count&#34;:0}"><button data-on:click__debounce.300ms="@post(&#39;/inc&#39;)" data-indicator:busy data-attr:disabled="$busy">+</button><span data-text="$count"></span></div>
```

- `Signals` JSON-encodes its argument, so maps and structs with `json` tags both work. It panics on a value that cannot be encoded; `MarshalSignals` returns the error instead, and `SignalsJSON` takes a hand-written object literal.
- Modifiers follow a `__` after the attribute name and are passed as trailing arguments, such as `Debounce`, `Throttle`, `Window`, `Prevent`, `Once`, `Outside` and `ViewTransition`. Modifiers with arguments chain further methods, for example `Debounce(300*time.Millisecond).Leading()`.
- `Get`, `Post`, `Put`, `Patch`, `Delete` and `Query` return an action expression as a quoted string, such as `@post('/inc')`; pass `RequestOptions{...}.String()` for the options object. `SetAll`, `ToggleAll`, `Peek` and the Pro `Clipboard`, `Fit` and `Intl` return the other action forms. These go in the expression argument of `OnClick`, `On` and the like.
- `Custom(plugin, key, value, mods...)` builds any Datastar attribute the package does not cover.

The package targets Datastar v1 and emits the v1 colon key syntax (`data-on:click`, `data-signals:name`). It was verified against the Datastar v1.0.4 reference ([attributes](https://data-star.dev/reference/attributes), [actions](https://data-star.dev/reference/actions)) on 2026-10-31, including the Pro attributes and actions, which are marked "Pro" in the doc comments. Names and modifiers change between Datastar releases, so re-check the reference when you upgrade.

## Escaping and Safety

- `Text` HTML-escapes its input, so user data is safe to pass to it. `TextUnsafe` writes its input verbatim and should only receive trusted markup.
- Attribute values are always escaped. This covers `Attr`, `Data`, `Aria`, the typed setters (such as `Title` or `Href`), and `Generic` attributes. `AttrUnsafe` and `DataUnsafe` skip escaping, so only use them with trusted values.
- `script` and `style` take raw text through `Text`, with no HTML escaping, so only pass trusted code. The one rewrite is that a `</script` or `</style` sequence (in any case) is changed to `<\/script` or `<\/style`, so the content cannot close the element early.
- URL attributes such as `href` and `src` are escaped but not scheme-filtered, so `Href("javascript:...")` renders as given. Pass untrusted URLs through `htemel.URL`, which returns `ErrUnsafeURL` for anything outside the policy: relative references, absolute `http`, `https`, `mailto`, `tel`, `sms`, `ftp` and `blob` URLs, and `data:` URLs for raster `image/*` types (not SVG). `htemel.URLOr` returns a fallback instead of an error, and `htemel.MustURL` panics, so use it only for constants.

```go
P().Text(`<img src=x onerror="alert(1)">`) // <p>&lt;img src=x onerror=&#34;alert(1)&#34;&gt;</p>

A().Href(htemel.URLOr(userInput, "#")).Text("Profile")
```

**Output** (for `userInput` set to `javascript:alert(1)`):

```html
<a href="#">Profile</a>
```

## Goals

The primary goal of **htemel** is to provide a **compile-time safe** method for generating HTML. HTML elements are represented as functions that return structs, enabling method chaining for attribute modifications. Each element has a restricted set of attributes, ensuring only valid attributes appear in code completion or editor hints. For enumerated attributes, types are generated to conform to the HTML specification.

**Example**:

```go
html.Meta().HttpEquiv(html.MetaHttpEquivContentType)
```

Attributes are mutually exclusive per element, with the most recent chain call taking precedence:

```go
// Space-separated token attributes (e.g., class) accept multiple strings
html.A().Class("active", "hover:bg-indigo-400").Class("this-takes-precedence")
```

**Output**:

```html
<a class="this-takes-precedence"></a>
```

## Why htemel?

After evaluating alternatives like `templ`, `gomponents`, and `gostar`, none fully met the need for a robust, type-safe, and extensible HTML generation library in Go:

- **templ**: Felt too much like writing HTML with interpolated templates.
- **gomponents**: Promising but too permissive with attributes, and distinguishing global vs. element-specific attributes was cumbersome.
- **gostar**: Showed potential but appeared unmaintained, with challenges in implementing custom components like TailwindPlus.

Thus, **htemel** was created to address these gaps with a focus on type safety and extensibility.

## Development

- `go generate ./...` regenerates the `html`, `svg` and `mathml` packages from `specs/html.json`, `specs/svg.json` and `specs/mathml.json`. Files in `html/`, `svg/` and `mathml/` are generated, so change the generator in `cmd/genelements` or the spec rather than editing them by hand.
- `make check-generated` regenerates the code and fails if `html/`, `svg/` or `mathml/` differ from what is committed, which catches a spec change without its output.
- `make test` runs `go test -race ./...`.
- `make lint` runs `golangci-lint` when it is installed, and falls back to `go vet` otherwise.

## Contributing

Contributions are welcome! Fork the repository and submit a pull request. Please align new features with the API's existing design and direction.

## Generators

**htemel** includes generators to streamline development:

1. **HTML Spec Generator**: Pulls from the [Living HTML Standard](https://html.spec.whatwg.org/). Note: The parser is a work in progress, with some hand-crafted attributes to fill gaps.
2. **Element Generator**: Generates namespaced elements from a provided spec. You can write custom specs for your web components and generate corresponding elements.

The `svg` and `mathml` packages are generated by the same tool from the specs maintained in [go-htemel/spec](https://github.com/go-htemel/spec). Contributions to enhance the generator are encouraged!
