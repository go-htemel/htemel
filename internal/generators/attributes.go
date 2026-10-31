package generators

import (
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/go-htemel/spec"
)

// attrDoc returns the doc comment for a setter: a Deprecated paragraph when the attribute is deprecated
// and, for URL-valued string attributes, a hint about validating untrusted input.
func attrDoc(name string, a spec.Attribute) string {
	var doc string
	if u, ok := a.(*spec.AttributeTypeString); ok && u.URL {
		doc += "// " + name + " sets the \"" + a.GetName() + "\" attribute.\n// Value is a URL; validate untrusted input with htemel.URL.\n"
	}
	if a.IsDeprecated() {
		if doc == "" {
			doc += "// " + name + " sets the \"" + a.GetName() + "\" attribute.\n"
		}
		doc += "//\n// Deprecated: the \"" + a.GetName() + "\" attribute is deprecated or obsolete in its source specification.\n"
	}
	return doc
}

func BuildAttributes(e *spec.Element) func() (*template.Template, ImportSet) {
	return func() (*template.Template, ImportSet) {
		imports := ImportSet{}

		addTypes := &bytes.Buffer{}
		buf := &bytes.Buffer{}

		for _, attr := range e.Attributes {
			start := buf.Len()
			switch a := attr.(type) {
			case *spec.AttributeTypeString:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(s string) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", s)
	
	return e
}
`)
			case *spec.AttributeTypeChar:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(r rune) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", string(r))
	
	return e
}
`)
			case *spec.AttributeTypeNumber:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(i int) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", strconv.Itoa(i))
	
	return e
}
`)
			case *spec.AttributeTypeFloat:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(f float64) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", strconv.FormatFloat(f, 'g', -1, 64))
	
	return e
}
`)
			case *spec.AttributeTypeBool:
				buf.WriteString(`
// ` + goIdent(a.Name) + ` sets the boolean attribute "` + a.Name + `". When b is true it renders bare;
// when false the attribute is omitted (and removed if previously set).
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(b bool) *` + goIdent(e.Tag) + `Element {
	e.attributes.SetBool("` + a.Name + `", b)
	
	return e
}
`)
			case *spec.AttributeTypeEnum:
				typeName := goIdent(e.Tag) + goIdent(a.Name)

				addTypes.WriteString("\ntype " + typeName + " string\n")

				addTypes.WriteString("\nconst (\n")
				allowedKeys := make([]string, 0, len(a.Allowed))
				for allowed := range a.Allowed {
					allowedKeys = append(allowedKeys, allowed)
				}
				sort.Strings(allowedKeys)
				for _, allowed := range allowedKeys {
					fixed := goIdent(allowed)
					addTypes.WriteString("\t" + typeName + fixed + " " + typeName + " = \"" + allowed + "\"\n")
				}

				// Generate function to allow empty attribute value
				if a.AllowEmpty {
					addTypes.WriteString("\t" + typeName + "Empty " + typeName + " = \"\"\n")
				}
				addTypes.WriteString(")\n")

				if a.AllowEmpty {
					// Variadic: no argument renders the bare attribute; with several, the last wins.
					buf.WriteString(`
// ` + goIdent(a.Name) + ` sets the "` + a.Name + `" attribute. Called with no argument, ` + goIdent(a.Name) + `() renders the
// bare attribute (empty value); with arguments the last one is used.
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(v ...` + typeName + `) *` + goIdent(e.Tag) + `Element {
	var val ` + typeName + `
	if len(v) > 0 {
		val = v[len(v)-1]
	}
	e.attributes.Set("` + a.Name + `", string(val))
	
	return e
}
`)
				} else {
					buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(a ` + typeName + `) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", string(a))
	
	return e
}
`)
				}
				if a.AllowCustom {
					// Generate func to return a custom attribute value
					buf.WriteString(`
func  ` + goIdent(e.Tag) + goIdent(a.Name) + `Custom(s string) ` + typeName + ` {
	return ` + typeName + `(s)
}
`)
				}
			case *spec.AttributeTypeSST:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(s ...string) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", strings.Join(s, " "))
	
	return e
}
`)
			case *spec.AttributeTypePrefixedCustom:
				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `Unsafe(name string, s string) *` + goIdent(e.Tag) + `Element {
	tag := strings.ToLower("` + a.Name + `-" + name)
	
	e.attributes.SetUnsafe(tag, s)
	
	return e
}

func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(name string, s string) *` + goIdent(e.Tag) + `Element {
	tag := strings.ToLower("` + a.Name + `-" + name)

	e.attributes.Set(tag, s)

	return e
}
`)
			}
			if doc := attrDoc(goIdent(attr.GetName()), attr); doc != "" {
				seg := buf.String()[start:] // begins with "\n"
				if strings.HasPrefix(seg[1:], "//") {
					doc += "//\n" // keep the setter's own comment out of the Deprecated paragraph
				}
				buf.Truncate(start)
				buf.WriteString("\n" + doc + seg[1:])
			}
		}

		out := &bytes.Buffer{}

		tmpl := template.Must(template.New("BuildAttributes").Parse(addTypes.String()))
		if err := tmpl.Execute(out, e); err != nil {
			panic(err)
		}
		tmpl = template.Must(tmpl.Parse(buf.String()))
		if err := tmpl.Execute(out, e); err != nil {
			panic(err)
		}

		tmpl = template.Must(tmpl.Parse(out.String()))

		return tmpl, imports.Add("strings").Add("strconv")
	}
}
