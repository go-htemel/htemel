# htemel Example Web Application

A simple web application demonstrating how to use **htemel** to build server-rendered HTML pages with Go's standard library.

## Overview

This example demonstrates:
- Using htemel to generate HTML views in pure Go
- Organizing views into reusable components
- Building a multi-page application with navigation
- Integrating Tailwind CSS for styling

## Project Structure

```
examples/webapp/
  main.go              # Application entry point and routing
  handlers/
    handlers.go      # HTTP request handlers
  views/
    layout.go        # Base layout and navigation components
    index.go         # Home page view
    about.go         # About page view
    contact.go       # Contact page view
```

## Running the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Features

### Layout System

The application uses a shared layout (`views.Layout`) that provides:
- Common HTML structure with DOCTYPE and meta tags
- Tailwind CSS integration via CDN
- Responsive navigation bar with active page highlighting
- An inline SVG logo icon, built with the `svg` package and colored with `currentColor`
- Dark mode support

### Routes

- `/` - Home page
- `/about` - About page
- `/contact` - Contact page

### Type-Safe Views

All views are written in pure Go using htemel's chainable API:

```go
func Index() FlowContent {
    return PageContent(
        "Home",
        P().Text("Welcome to the htemel example web application!"),
    )
}
```

## Key Concepts

### Rendering Views

Views are rendered to a buffer and then written to the HTTP response:

```go
buf := &bytes.Buffer{}
if err := view.Render(buf); err != nil {
    http.Error(w, "Failed to render view", http.StatusInternalServerError)
    return
}

w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Write(buf.Bytes())
```

### Component Composition

Components can be nested and reused. The `Layout` function accepts flow content, because its
children end up inside a `<div>` in `<body>`; page views therefore return `FlowContent` rather than
a plain `Node`, and the compiler rejects a view that could not legally go there:

```go
Layout(PageData{Title: "Home", Request: r}, Index())
```

### Repeated Content

The navigation renders its links with `Each`, which calls a function for each item at render time:

```go
Each(links, func(link navLink) Node {
    link.active = currentPage == link.href
    return NavigationLink(link)
})
```

### Inline SVG

`LogoIcon` in `views/layout.go` is an SVG icon built with the `svg` package. It is imported under its package name because `svg` shares names with the `html` package. The icon uses `currentColor`, so it follows the text color of the link around it.

### Dynamic Styling

The navigation uses conditional classes to highlight the active page:

```go
ToggleClass("text-white border-red-500/70", link.active)
```

## Extending the Example

To add a new page:

1. Create a view function in `views/`
2. Add a handler in `handlers/handlers.go`
3. Register the route in `main.go`
4. Optionally add it to the navigation links in `views/layout.go`
