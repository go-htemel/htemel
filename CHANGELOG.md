# Changelog

## v0.4.0 — 2026-10-31

Breaking release. It implements ADRs 0008, 0009, 0010 and 0011 and settles 0013 (rejected, no
helper). The generated `html` package goes from 6,651 to 3,985 top-level `func` declarations, and
`*html.DivElement` from 54 methods to 32 (plus the four child-interface marker methods). `svg` and
`mathml` together go from 7,510 to 7,134.

> **Behaviour change you will not get a compile error for: `Class` now adds classes instead of
> replacing them.** `Div().Class("a").Class("b")` used to render `class="b"`; it now renders
> `class="a b"`. Search your code for a second `.Class(` on the same element, or a `.Class(` on an
> element returned by a component, and switch the ones that meant "replace" to `.SetClass(`.

### Removed: per-element conditionals (ADR 0008)

The package-level `<Tag>If(cond, children...)` and `<Tag>Ternary(cond, then, otherwise)`
constructors and the element methods `If(cond, child)`, `IfElse(cond, then, else)` and
`IfThen(cond, fn)` are gone from `html`, `svg` and `mathml`. A survey of the `ui` library found no
call sites of the element methods, so `.If` went too. `htemel.If`, `htemel.IfElse`, `Each`,
`ForEach` and `.With` remain. Elements no longer carry a `skipRender` flag.

| Before | After |
|---|---|
| `html.DivIf(cond, kids...)` | `htemel.If(cond, html.Div(kids...))` |
| `html.DivTernary(cond, a, b)` | `html.Div(htemel.IfElse(cond, a, b))` |
| `el.If(cond, child)` | `el.Children(htemel.If(cond, child))`, or an `if` / `.With` to keep the child type-checked |
| `el.IfElse(cond, a, b)` | `el.Children(htemel.IfElse(cond, a, b))` |
| `el.IfThen(cond, fn)` | `el.With(func(e *html.DivElement) { if cond { fn(e) } })`, or a plain `if` |

`htemel.If`/`IfElse` return a wildcard, so the chosen child is no longer checked against the
parent's content model; use `if` or `.With` where you want that check.

### Changed: `Class` is additive (ADR 0009)

- `Class(classes...)` appends to the existing class attribute. Each argument may hold several
  space-separated classes; a class that is already present is skipped, order is preserved, and
  whitespace is normalised. `Class()` or `Class("")` on an element without classes adds nothing.
- New `SetClass(classes...)` replaces the attribute (de-duplicated; removes it when empty).
- `AddClass` is now a deprecated alias of `Class` and will be removed in a later release.
  `ToggleClass` is unchanged.
- Setting `class` with `Attr`, `Attrs(htemel.Attr("class", ...))`, `SetAttr` or the `datastar`
  helpers still writes the value as given (replacing), as before.
- `class` is no longer generated from the spec's attribute list but as a built-in, on every
  element of every generated package (`CheckNames` reserves `Class` and `SetClass`).

Migration: where code relied on `Class` replacing earlier classes, use `SetClass`. Calls of
`AddClass(...)` can be renamed to `Class(...)`.

### Changed: rare global attributes move into `html.Globals` (ADR 0010)

Every html element keeps these global attributes as direct methods: `Id`, `Class`/`SetClass`/
`ToggleClass`, `Style`, `Title`, `Lang`, `Dir`, `Hidden`, `Tabindex`, `Role`, `Autofocus`,
`Draggable`, `Popover`, `Nonce`, `Data`/`DataUnsafe`, `Aria`/`AriaUnsafe` (and the generic `Attr`,
`AttrUnsafe`, `BoolAttr`, `Attrs`, `SetAttr`, `SetAttrUnsafe`). Element-specific attributes such as
`Popovertarget` are untouched. The other 21 global attributes are fields of the new
`html.Globals` struct, applied with the new `Globals(g html.Globals)` method on every element:

| Before | After |
|---|---|
| `Div().Itemprop("x")` | `Div().Globals(html.Globals{Itemprop: []string{"x"}})` |
| `Div().Itemscope(true).Itemtype(t)` | `Div().Globals(html.Globals{Itemscope: true, Itemtype: []string{t}})` |
| `Div().Itemid(u)` / `.Itemref(ids...)` | `Globals{Itemid: u}` / `Globals{Itemref: ids}` |
| `Input().Enterkeyhint(html.EnterkeyhintSend)` | `Input().Globals(html.Globals{Enterkeyhint: html.EnterkeyhintSend})` |
| `.Inputmode(m)`, `.Autocapitalize(a)`, `.Autocorrect(a)` | `Globals{Inputmode: m, Autocapitalize: a, Autocorrect: a}` |
| `.Spellcheck(v)`, `.Translate(v)`, `.Writingsuggestions(v)` | `Globals{Spellcheck: v, Translate: v, Writingsuggestions: v}` |
| `.Contenteditable(v)` | `Globals{Contenteditable: v}` |
| `.Inert(true)`, `.Headingreset(true)` | `Globals{Inert: true, Headingreset: true}` |
| `.Headingoffset(n)` | `Globals{Headingoffset: n}` |
| `.Accesskey('k')` | `Globals{Accesskey: "k"}` (a string now, not a rune) |
| `.Slot(s)`, `.Is(s)`, `.Exportparts(s)` | `Globals{Slot: s, Is: s, Exportparts: s}` |
| `.Part(p...)` | `Globals{Part: p}` |

Fields keep the shared enum types (`html.Translate`, `html.Spellcheck`, ...). Zero values are
skipped (so `Inert: false` does not remove an attribute set earlier, and an empty-valued enum such
as `Spellcheck()` with no argument now needs `.Attr("spellcheck", "")`); bools render bare; slices
are joined with spaces; fields are applied in declaration order, so output stays deterministic.
`svg` and `mathml` keep every attribute as a method.

Generator: `generators.Options.DirectGlobals` and `generators.HTMLDirectGlobals` (the keep-list);
`cmd/genelements -globals-struct` turns it on, and the html `go:generate` line passes it. The
generator writes `html/globals.go`, and `CheckNames`/`Plan.Check` reserve `Globals`.

### Added: `RenderContext`, `ContextOf`, `Write` and `Handler` (ADR 0011)

- `htemel.RenderContext(ctx, w, n) error` renders through the same pooled buffer as `Render` and
  makes `ctx` available to every component through `htemel.ContextOf(w) context.Context`. If
  `ctx` is done before or after rendering, the buffered output is discarded and `ctx.Err()`
  returned. `ContextOf` returns `context.Background()` for any other writer, so plain `Render`
  behaves as `RenderContext(context.Background(), ...)`; a nested `htemel.Render` inside a
  context render keeps the context.
- `htemel.Write(w http.ResponseWriter, r *http.Request, n Node) error` renders the whole page into
  a buffer (with `r.Context()` available through `ContextOf`), sets
  `Content-Type: text/html; charset=utf-8` unless one is set, and writes it in one call. On a
  rendering error it writes **nothing** (no header, status or body) and returns the error, so the
  caller picks the response.
- `htemel.Handler(func(r *http.Request) Node) http.Handler` serves the page with `Write`: 200 on
  success; on a rendering error it logs (standard `log`) and sends a plain-text 500 with no partial
  page.
- `Node.Render` is unchanged. Its doc comment, the generated `Render` docs and `htemel.Render` now
  explain that a failed render can leave partial output in `w` (a half page with a 200 status on
  an `http.ResponseWriter`) and point to `Write`/`Handler`.
- The webapp example uses `Handler` and `Write`.

Migration: replace hand-written "render to a `bytes.Buffer`, set the header, write" handlers, and
any `htemel.Render(w, page)` / `page.Render(w)` on an `http.ResponseWriter`, with `htemel.Write` or
`htemel.Handler`.

### Not added: composite helper (ADR 0013, rejected)

A `Composite`/`Delegating` embedding helper was prototyped and not shipped. It cannot be embedded
next to the outer element (`*html.DivElement`): the promoted `Render`/`SetAttr` become ambiguous
and the type stops implementing `htemel.Element`. Used instead of the outer element, it loses the
wrapper's builder methods and still needs a typed `Attrs` and `Control` per composite, saving
about two lines. The README documents the three-method pattern (embed the outer element, declare
`Attrs`, `SetAttr` and `SetAttrUnsafe` that forward to the control).

## v0.3.0 — 2026-10-31

Breaking release. Highlights: the HTML content model is checked at compile time (`P(Div())`,
`Ul(Div())` and `Img().Children(...)` no longer compile), and enum attribute types are shared per
attribute (`html.Popover` instead of 112 `<El>Popover` types; the html package goes from 1331
enum types to 40, svg from 1385 to 61, mathml from 73 to 11).

Generated from go-htemel/spec with content models (the `categories` and `content_model` element
fields, from the WHATWG element index).

### Content models

- New in package `htemel`: the category interfaces `FlowContent`, `PhrasingContent` (which is
  also a `FlowContent`), `MetadataContent`, `SectioningContent`, `HeadingContent`,
  `EmbeddedContent`, `InteractiveContent`, `PalpableContent`, `ScriptSupportingContent` and
  `TextContent`; the marker structs `FlowMarker`, `PhrasingMarker`, ... `TextMarker` that types
  embed to join a category; `Wildcard`, which satisfies every content interface; `NodeFunc`, a
  function that renders a node (also a wildcard); and `Unchecked(n Node) NodeFunc`.
- Every generated element embeds the markers of its categories, so `html.Span()` is a
  `htemel.PhrasingContent` and `html.Div()` a `htemel.FlowContent`.
- Elements whose model lists tags get a child interface in their package, implemented by the
  permitted elements: `html.UlChild`, `OlChild`, `MenuChild`, `DlChild`, `HtmlChild`,
  `TableChild`, `TbodyChild`, `TheadChild`, `TfootChild`, `TrChild`, `ColgroupChild`,
  `SelectChild`, `OptgroupChild`, `OptionChild`, `DatalistChild`, `PictureChild`, `HgroupChild`,
  `FigureChild`, `FieldsetChild`, `DetailsChild`, `LegendChild`, `SummaryChild`, `RubyChild`,
  `AudioChild`, `VideoChild`. Their marker methods (`HTMLUlChild()`, ...) are exported and have
  no behaviour.
- `svg` and `mathml` get a closed, loose model: every element accepts `...svg.Content` /
  `...mathml.Content`, implemented by every element of the package and by text. `svg.Svg()` and
  `mathml.Math()` are embedded, phrasing, flow and palpable content, so they still nest in HTML;
  other SVG/MathML elements cannot be placed in HTML directly. `svg.ForeignObject` and
  `mathml.AnnotationXml` accept any `htemel.Node`.
- Not checked: transparent elements (`a`, `ins`, `del`, `map`, `object`, `canvas`, `slot`,
  `noscript`, `audio`, `video`) accept any flow content; descendant, ordering and cardinality
  rules; `template`, which accepts any node; and dynamic content (see below).
- New files generated into package `htemel`: `slots_html.go`, `slots_svg.go` and
  `slots_mathml.go`, which add the child-interface marker methods to `Wildcard`, `NodeFunc`,
  text and the category markers. `cmd/genelements` has a new `-slots` flag that writes them.

### Migration

- **Children parameters are typed.** `Div(children ...htemel.Node)` is now
  `Div(children ...htemel.FlowContent)`; likewise `DivIf`, `DivTernary(cond, then, otherwise)`,
  `.Children(...)`, `.If(cond, child)` and `.IfElse(cond, a, b)` on every element, each with that
  element's child type (`htemel.PhrasingContent` for `P`, `html.UlChild` for `Ul`,
  `htemel.TextContent` for `Title`, `htemel.MetadataContent` for `Head`, ...). `.With`, `.Attrs`,
  `.IfThen` and the attribute setters are unchanged.
  - Code that puts block content in phrasing content (a `Div` in a `P`, `Span`, `Button`, `A`'s
    phrasing parent, `H1`...) no longer compiles. That markup was invalid HTML (browsers silently
    restructure a `<div>` inside a `<p>`); use a flow container instead.
  - Components that return or accept `htemel.Node` must use the content type instead
    (`func Card(children ...htemel.FlowContent) htemel.FlowContent`), or wrap the value with
    `htemel.Unchecked(n)`.
  - A custom `Node` type passed as a child needs a marker: embed `htemel.FlowMarker`,
    `htemel.PhrasingMarker`, ... (or `htemel.Wildcard` to accept it anywhere).
- **Dynamic helpers return `htemel.NodeFunc`.** `Each`, `EachIndex`, `ForEach`, `ForEach2`, `If`
  and `IfElse` returned `htemel.Node`; they now return `htemel.NodeFunc`, which is accepted as a
  child of any element. `Group`, `Generic`, `GenericVoid` and `Doctype` keep their result types,
  which now embed `htemel.Wildcard`. A variable declared `var n htemel.Node = htemel.Each(...)`
  loses that ability; keep the `NodeFunc` or wrap with `Unchecked`.
- **Text only where text is permitted.** `Text` and `Textf` are removed from elements whose
  content model has no text: `Colgroup`, `Dl`, `Head`, `Hgroup`, `Html`, `Menu`, `Ol`,
  `Optgroup`, `Picture`, `Select`, `Table`, `Tbody`, `Tfoot`, `Thead`, `Tr`, `Ul`.
- **No children on elements with an empty content model.** `Iframe` and `Selectedcontent` take an
  end tag but no content: their children parameter and the `Children`, `If`, `IfElse`, `Text`,
  `Textf` and `<Tag>Ternary` functions are gone (void elements never had them).
- **Shared enum types.** Each enumerated attribute now has one type named after the attribute,
  shared by every element: `html.DivPopover` / `html.SpanPopover` / ... become `html.Popover`,
  constants `html.DivPopoverManual` become `html.PopoverManual`, `html.DivHiddenUntilFound`
  becomes `html.HiddenUntilFound`, `html.ADownloadCustom(s)` becomes `html.DownloadCustom(s)`,
  `svg.PathStrokeLinecapRound` becomes `svg.StrokeLinecapRound`, and so on: drop the element
  prefix. This also applies to attributes that exist on a single element (`html.TrackKind` →
  `html.Kind`, `html.ThScope` → `html.Scope`, `html.MetaCharset` → `html.Charset`).
  Attributes whose allowed values differ between elements keep per-element types:
  `html.ButtonType`, `html.InputType`, `html.ScriptType`; `svg.AnimateTransformType`,
  `svg.FeColorMatrixType`, `svg.FeFuncAType`/`B`/`G`/`R`, `svg.FeTurbulenceType`,
  `svg.FeCompositeOperator`, `svg.FeMorphologyOperator`. The list is at the top of each
  package's generated `enums.go`. AllowEmpty setters stay variadic (`Span().Popover()`).

## v0.2.0 — 2026-10-31

Highlights: every attribute value is now escaped and attribute order is deterministic; boolean attributes render correctly; rendering is ~4x faster with `htemel.Render`; new `svg`, `mathml` and `datastar` packages; `Each`/`ForEach`/`If`, `URL`, `Attribute`/`Attrs` helpers; variadic `Hidden()`-style setters; CI, lint, fuzz and golden tests.

Generated packages now come from [go-htemel/spec v0.3.0](https://github.com/go-htemel/spec), whose html and svg data is built by `specgen2` from webref, browser-compat-data and the WHATWG index tables instead of hand-written lists. New `mathml` package (31 elements). The setter changes below follow from corrected attribute types; removed setters were attributes that do not exist in the standards. `svg.Svg().Xmlns(...)` is gone because the namespace declaration is ignored in HTML; use `.Attr("xmlns", ...)` for standalone SVG files.


## html

### Changed setters

- `Autocomplete(string)` -> `Autocomplete(...string)` on 2 elements (Select, Textarea)
- `Blocking(<El>Blocking)` -> `Blocking(...string)` on 2 elements (Link, Script)
- `Blocking(string)` -> `Blocking(...string)` on 1 elements (Style)
- `Charset(string)` -> `Charset(<El>Charset)` on 1 elements (Meta)
- `Download(bool)` -> `Download(...<El>Download)` on 2 elements (A, Area)
- `Enctype(string)` -> `Enctype(<El>Enctype)` on 1 elements (Form)
- `For(string)` -> `For(...string)` on 1 elements (Output)
- `Formenctype(string)` -> `Formenctype(<El>Formenctype)` on 2 elements (Button, Input)
- `Max(int)` -> `Max(string)` on 1 elements (Input)
- `Min(int)` -> `Min(string)` on 1 elements (Input)
- `Popover(string)` -> `Popover(...<El>Popover)` on 112 elements
- `Referrerpolicy(string)` -> `Referrerpolicy(...<El>Referrerpolicy)` on 6 elements (A, Area, Iframe, Img, Link, Script)
- `Rows(string)` -> `Rows(int)` on 1 elements (Textarea)
- `Scope(string)` -> `Scope(<El>Scope)` on 1 elements (Th)
- `Type(string)` -> `Type(<El>Type)` on 1 elements (Script)

### Removed setters

- `Playsinline(bool)` removed from 1 elements (Audio)
- `Shadowrootdelegatefocus(bool)` removed from 1 elements (Template)
- `Span(int)` removed from 2 elements (Td, Th)

### Added setters

- `Accesskey` added to 1 elements (Sup)
- `AddClass` added to 1 elements (Sup)
- `Aria` added to 1 elements (Sup)
- `AriaUnsafe` added to 1 elements (Sup)
- `Attr` added to 1 elements (Sup)
- `AttrUnsafe` added to 1 elements (Sup)
- `Attrs` added to 1 elements (Sup)
- `Autocapitalize` added to 1 elements (Sup)
- `Autocorrect` added to 1 elements (Sup)
- `Autofocus` added to 1 elements (Sup)
- `BoolAttr` added to 1 elements (Sup)
- `Capture` added to 1 elements (Input)
- `Children` added to 1 elements (Sup)
- `Class` added to 1 elements (Sup)
- `Colspan` added to 2 elements (Td, Th)
- `Contenteditable` added to 1 elements (Sup)
- `Controls` added to 1 elements (Img)
- `Controlslist` added to 2 elements (Audio, Video)
- `Data` added to 1 elements (Sup)
- `DataUnsafe` added to 1 elements (Sup)
- `Dir` added to 1 elements (Sup)
- `Disablepictureinpicture` added to 1 elements (Video)
- `Disableremoteplayback` added to 2 elements (Audio, Video)
- `Draggable` added to 1 elements (Sup)
- `Enterkeyhint` added to 1 elements (Sup)
- `Exportparts` added to 113 elements
- `For` added to 1 elements (Template)
- `Headingoffset` added to 113 elements
- `Headingreset` added to 113 elements
- `Hidden` added to 1 elements (Sup)
- `Hreflang` added to 1 elements (Area)
- `Id` added to 1 elements (Sup)
- `If` added to 1 elements (Sup)
- `IfElse` added to 1 elements (Sup)
- `IfThen` added to 1 elements (Sup)
- `Inert` added to 1 elements (Sup)
- `Inputmode` added to 1 elements (Sup)
- `Is` added to 113 elements
- `Itemid` added to 1 elements (Sup)
- `Itemprop` added to 1 elements (Sup)
- `Itemref` added to 1 elements (Sup)
- `Itemscope` added to 1 elements (Sup)
- `Itemtype` added to 1 elements (Sup)
- `Lang` added to 1 elements (Sup)
- `Loading` added to 2 elements (Audio, Video)
- `Nonce` added to 1 elements (Sup)
- `Part` added to 113 elements
- `Popover` added to 1 elements (Sup)
- `Role` added to 1 elements (Sup)
- `Shadowrootdelegatesfocus` added to 1 elements (Template)
- `Shadowrootslotassignment` added to 1 elements (Template)
- `Slot` added to 1 elements (Sup)
- `Spellcheck` added to 1 elements (Sup)
- `Style` added to 1 elements (Sup)
- `Tabindex` added to 1 elements (Sup)
- `Text` added to 1 elements (Sup)
- `Textf` added to 1 elements (Sup)
- `Title` added to 1 elements (Sup)
- `ToggleClass` added to 1 elements (Sup)
- `Translate` added to 1 elements (Sup)
- `Type` added to 1 elements (Area)
- `Webkitdirectory` added to 1 elements (Input)
- `With` added to 1 elements (Sup)
- `Writingsuggestions` added to 1 elements (Sup)

## svg

### Changed setters

- `Dx(string)` -> `Dx(float64)` on 2 elements (FeDropShadow, FeOffset)
- `Dy(string)` -> `Dy(float64)` on 2 elements (FeDropShadow, FeOffset)
- `Fill(string)` -> `Fill(<El>Fill)` on 4 elements (Animate, AnimateMotion, AnimateTransform, Set)
- `PathLength(string)` -> `PathLength(float64)` on 7 elements
- `Ping(string)` -> `Ping(...string)` on 1 elements (A)
- `PreserveAlpha(bool)` -> `PreserveAlpha(<El>PreserveAlpha)` on 1 elements (FeConvolveMatrix)

### Removed setters

- `Href(string)` removed from 1 elements (Filter)
- `Path(string)` removed from 1 elements (TextPath)
- `XlinkHref(string)` removed from 4 elements (Animate, AnimateMotion, AnimateTransform, Set)
- `Xmlns(string)` removed from 1 elements (Svg)

### Added setters

- `BaseProfile` added to 1 elements (Svg)
- `BaselineShift` added to 63 elements
- `Clip` added to 63 elements
- `FontSizeAdjust` added to 63 elements
- `FontStretch` added to 63 elements
- `FontVariant` added to 63 elements
- `GlyphOrientationHorizontal` added to 63 elements
- `GlyphOrientationVertical` added to 63 elements
- `Origin` added to 1 elements (AnimateMotion)
- `RequiredExtensions` added to 63 elements
- `SystemLanguage` added to 25 elements
- `TextOverflow` added to 63 elements
- `Transform` added to 44 elements
- `TransformOrigin` added to 44 elements
- `XlinkActuate` added to 1 elements (A)
- `XlinkShow` added to 1 elements (A)
- `XlinkTitle` added to 1 elements (A)
