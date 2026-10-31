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

- Use `With` to build repeated content from a loop. It passes the typed element to a callback, so the loop can add children to it:

  ```go
  func ItemList(items []string) Node {
      return Ul().Class("list").With(func(ul *UlElement) {
          for _, item := range items {
              ul.Children(Li().Text(item))
          }
      })
  }
  ```

- Use `Group` to return several sibling nodes without a wrapping element.

The [webapp example](examples/webapp/views/layout.go) shows these patterns in a complete application: a shared layout, a navigation component built with `With`, and a page wrapper that accepts children.

## Escaping and Safety

- `Text` HTML-escapes its input, so user data is safe to pass to it. `TextUnsafe` writes its input verbatim and should only receive trusted markup.
- Attribute values are always escaped. This covers `Attr`, `Data`, `Aria`, the typed setters (such as `Title` or `Href`), and `Generic` attributes. `AttrUnsafe` and `DataUnsafe` skip escaping, so only use them with trusted values.
- `script` and `style` take raw text through `Text`, with no HTML escaping, so only pass trusted code. The one rewrite is that a `</script` or `</style` sequence (in any case) is changed to `<\/script` or `<\/style`, so the content cannot close the element early.
- URL attributes such as `href` and `src` are escaped but not scheme-filtered. `Href("javascript:...")` renders as given. Validate untrusted URLs (for example, allow only `http`, `https`, or relative paths) before passing them to `Href`.

```go
P().Text(`<img src=x onerror="alert(1)">`) // <p>&lt;img src=x onerror=&#34;alert(1)&#34;&gt;</p>
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

- `go generate ./...` regenerates the `html` package from `specs/html.json`. Files in `html/` are generated, so change the generator or the spec rather than editing them by hand.
- `go test ./...` runs the test suite.

## Contributing

Contributions are welcome! Fork the repository and submit a pull request. Please align new features with the API's existing design and direction.

## Generators

**htemel** includes generators to streamline development:

1. **HTML Spec Generator**: Pulls from the [Living HTML Standard](https://html.spec.whatwg.org/). Note: The parser is a work in progress, with some hand-crafted attributes to fill gaps.
2. **Element Generator**: Generates namespaced elements from a provided spec. You can write custom specs for your web components and generate corresponding elements.

Support for additional markup languages like SVG or MathML is planned. Contributions to enhance the generator are encouraged!
