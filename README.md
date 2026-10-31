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

func MainLayout(children ...FlowContent) Node {
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

func Navigation(menuItems ...UlChild) FlowContent {
    return Nav(
        // Child elements can be added using the Children method
        Ul().Class("flex list-none").Children(menuItems...),
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

### Content Models

Every element only accepts the children its specification permits, and the compiler enforces it.
`Div` takes flow content, `P` and `Span` take phrasing content, `Ul` takes `<li>` and
script-supporting elements, `Title` and `Script` take text, and void elements such as `Img` have
no `Children` method at all:

```go
P(Span().Text("ok"))       // compiles: span is phrasing content
P(Div())                   // does not compile: *DivElement does not implement htemel.PhrasingContent
Ul(Li().Text("ok"))        // compiles
Ul(Div())                  // does not compile: missing method HTMLUlChild
Img().Children(Span())     // does not compile: Children undefined
Ul().Text("x")             // does not compile: text is not permitted in <ul>
```

How the checks are expressed:

- **Categories.** Package `htemel` has an interface per content category used by the content
  models: `FlowContent`, `PhrasingContent`, `MetadataContent`, `SectioningContent`,
  `HeadingContent`, `EmbeddedContent`, `InteractiveContent`, `PalpableContent` and
  `ScriptSupportingContent`, plus `TextContent` for text nodes. Elements implement the interfaces
  of their categories (`Span()` is phrasing and flow content, `Div()` only flow content), and a
  `PhrasingContent` is always a `FlowContent`. Elements whose model is a single category take it
  directly: `Div(children ...htemel.FlowContent)`, `Head(children ...htemel.MetadataContent)`.
- **Element-specific children.** Where a model lists tags, the element package declares a child
  interface implemented by those tags and by the members of any listed category: `UlChild`
  (`<li>`, `<script>`, `<template>`), `HtmlChild` (`<head>`, `<body>`), `TableChild`, `TrChild`,
  `SelectChild`, `FigureChild` (`<figcaption>` and flow content), and so on.
- **Text.** `Text`, `Textf` and `htemel.Text(...)` are accepted wherever text is permitted.
  `Title`, `Script`, `Style`, `Textarea` and `Rp` take only text (`htemel.TextContent`).
- **No children.** Void elements, and `Iframe` and `Selectedcontent` (which have an end tag but
  no content), have no children parameter and no `Children`, `If`, `IfElse` or `Text` methods.

What is not checked:

- **Transparent elements** (`A`, `Ins`, `Del`, `Map`, `Object`, `Canvas`, `Slot`, `Noscript`,
  `Audio`, `Video`) take their parent's content model, which a Go type cannot follow. They accept
  any flow content and keep their own categories, so `A` is phrasing content and fits in a `P`,
  but `P(A(Div()))` compiles although the result is invalid. Check those yourself.
- **Descendant rules**, ordering and cardinality: no interactive content inside `<a>` or
  `<button>`, no `<form>` inside `<form>`, one `<title>` in `<head>`, `<summary>` first in
  `<details>`, attribute-dependent categories, and the like.
- **`Template`** accepts any `htemel.Node`: its contents are a separate document fragment that may
  hold anything, including `<tr>` or `<li>`.
- **Dynamic content.** `Group`, `Each`, `EachIndex`, `ForEach`, `ForEach2`, `If`, `IfElse`,
  `Generic`, `GenericVoid` and `Doctype` return nodes that are accepted as a child of any element
  (they embed `htemel.Wildcard` or are a `htemel.NodeFunc`), and the nodes inside them are not
  checked against the parent. Wrap a plain `Node` from elsewhere with `htemel.Unchecked(n)` to
  place it anywhere.
- **SVG and MathML** use a closed but loose model: every `svg` element accepts `...svg.Content`,
  implemented by every `svg` element and by text, and likewise for `mathml.Content`. The root
  `svg.Svg()` and `mathml.Math()` are embedded, phrasing and flow content, so they go wherever
  HTML permits them; other SVG and MathML elements cannot be placed in HTML directly.
  `svg.ForeignObject` and `mathml.AnnotationXml` accept any `htemel.Node`.

The categories and models come from the WHATWG element index, via
[go-htemel/spec](https://github.com/go-htemel/spec). The marker methods are compile-time only:
they have no fields and no effect on rendering speed.

### Writing Components

A component is a plain Go function. Arguments are ordinary Go values. Declare children and the
result with the content type they really are, so the compiler can check how the component is
used:

```go
func Card(title string, children ...FlowContent) FlowContent {
    return Div().Class("card").Children(
        H2().Text(title),
        Div().Class("card-body").Children(children...),
    )
}

func Badge(label string) PhrasingContent {
    return Span().Class("badge").Text(label)
}

P(Text("Status: "), Badge("new"))  // compiles: Badge is phrasing content
P(Card("Oops"))                    // does not compile: Card is flow content, not phrasing
```

Returning a plain `Node` still works, but such a value can only be passed where any node is
accepted (`Group`, `Render`, `Template`) unless you wrap it with `Unchecked`. A component type of
your own (a struct with a `Render` method) joins a category by embedding the marker struct, such
as `htemel.PhrasingMarker` or `htemel.FlowMarker`; embedding `htemel.Wildcard` opts it out of
checking.

- Use `Each` to build repeated content. It renders a function for each item at render time, so no intermediate slice of nodes is built:

  ```go
  func ItemList(items []string) FlowContent {
      return Ul().Class("list").Children(
          Each(items, func(item string) Node {
              return Li().Text(item)
          }),
      )
  }
  ```

  `EachIndex` also passes the index, `ForEach` and `ForEach2` range over an `iter.Seq` or `iter.Seq2` such as `slices.Values` or `maps.All`, and `If` and `IfElse` choose between nodes without a separate `if` block. Their results are accepted anywhere, so what `fn` returns is not checked against the parent (here, nothing stops `fn` from returning a `Div` inside the `Ul`).

- Use `Group` to return several sibling nodes without a wrapping element. A `Group` is also accepted anywhere and its children are not checked.

- The element-level `If`, `IfElse` and `<Tag>If`/`<Tag>Ternary` constructors are checked: `Ul().If(cond, Li())` compiles, `Ul().If(cond, Div())` does not.

`With` is still available when you need the typed element itself inside a function. The [webapp example](examples/webapp/views/layout.go) shows these patterns in a complete application: a shared layout, a navigation component built with `Each`, an inline SVG icon, and a page wrapper that accepts children.

#### Forwarding Attributes

Every generated element has an `Attrs` method that takes `htemel.Attribute` values, so a component can accept attributes from its caller and pass them on. `Attr` and `AttrUnsafe` build an attribute, and `If` returns the attribute only when the condition holds:

```go
func Button(label string, attrs ...htemel.Attribute) htemel.PhrasingContent {
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

### Enumerated Attributes

Attributes with a fixed set of values take a typed string. There is one type per attribute, named
after it and shared by every element that has it, with a constant per value:

```go
Div().Popover(PopoverManual).Dir(DirRtl)
Span().Popover()                  // AllowEmpty attributes are variadic: no argument renders the bare attribute
Img().Loading(LoadingLazy).Crossorigin(CrossoriginAnonymous)
A().Download(DownloadCustom("report.pdf")) // attributes that also take free-form values have a Custom constructor
```

When the allowed values of an attribute differ between elements, each element keeps its own type,
named `<Element><Attribute>`: `ButtonTypeSubmit`, `InputTypeEmail`, `ScriptTypeModule`. The list
is at the top of the generated `html/enums.go` (and `svg/enums.go`).

## SVG

The `svg` package generates the SVG elements, with tag and attribute names kept in their SVG case (`viewBox`, `linearGradient`, `feGaussianBlur`). The root `svg.Svg()` is embedded and phrasing content, so it nests directly inside HTML elements such as `Div`, `Button` or `Span`; inside it, every SVG element accepts any other (`...svg.Content`), and text.

Import `svg` under its package name rather than with a `.` import. It shares names with both `htemel` (`Text`) and `html` (`A`, `Title`, `Style`, `Script`, `Image`), so a dot import would conflict.

```go
func PlusIcon() htemel.PhrasingContent {
    return svg.Svg().
        ViewBox("0 0 24 24").
        Attr("width", "24").Attr("height", "24").
        Children(
            svg.Path().
                D("M4 12h16M12 4v16").
                Fill("none").
                Stroke("currentColor").
                StrokeWidth("2").
                StrokeLinecap(svg.StrokeLinecapRound),
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
