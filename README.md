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
    MainLayout(
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
    ).Render(os.Stdout)
}
```

**Output**:

```html
<!DOCTYPE html>
<html class="h-dvh bg-gray-200" lang="en">
<head>
    <meta charset="UTF-8">
    <meta content="width=device-width, initial-scale=1.0">
    <title>
        Example htemel Page
    </title>
    <link href="site.css" rel="stylesheet">
</head>
<body id="app-content">
<div>
    <nav id="main-navigation">
        <ul class="flex list-none">
        </ul>
    </nav>
    <div>
        <h4 class="text-xl font-semibold">
            Logged In
        </h4>
        <p>
            Welcome back!
        </p>
    </div>
</div>
</body>
</html>
```

### Controlling Element Placement

Chained calls respect the order of elements in the document, for example:

```go
package main

import (
    "os"
	
    . "github.com/go-htemel/htemel/html"
)

func main() {
    div := html.Div().
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
<div>
    Start Text<span>Spanned Text</span>End Text
</div>
```

## Goals

The primary goal of **htemel** is to provide a **compile-time safe** method for generating HTML. HTML elements are represented as functions that return structs, enabling method chaining for attribute modifications. Each element has a restricted set of attributes, ensuring only valid attributes appear in code completion or editor hints. For enumerated attributes, types are generated to conform to the HTML specification.

**Example**:

```go
html.Link().HttpEquiv(html.MetaHttpEquivContentType)
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

## Contributing

Contributions are welcome! Fork the repository and submit a pull request. Please align new features with the API's existing design and direction.

## Generators

**htemel** includes generators to streamline development:

1. **HTML Spec Generator**: Pulls from the [Living HTML Standard](https://html.spec.whatwg.org/). Note: The parser is a work in progress, with some hand-crafted attributes to fill gaps.
2. **Element Generator**: Generates namespaced elements from a provided spec. You can write custom specs for your web components and generate corresponding elements.

Support for additional markup languages like SVG or MathML is planned. Contributions to enhance the generator are encouraged!
