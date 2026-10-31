# Changelog

## Unreleased

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
