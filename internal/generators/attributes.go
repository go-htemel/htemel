package generators

import (
	"bytes"
	"sort"
	"text/template"

	"github.com/go-htemel/spec"
)

// urlAttributes lists attributes whose values are URLs. The spec (go-htemel/spec v0.1.1) has no URL
// attribute type, so these are currently generated as plain (HTML-escaped) strings. The list exists so
// scheme filtering (e.g. rejecting "javascript:") can be added later in one place.
var urlAttributes = map[string]bool{
	"href": true, "src": true, "action": true, "formaction": true, "poster": true, "cite": true,
	"data": true, "usemap": true, "itemid": true, "ping": true, "manifest": true,
}

// isURLAttr reports whether the named attribute carries a URL value.
func isURLAttr(name string) bool { return urlAttributes[name] }

var _ = isURLAttr // reserved for future URL sanitization

func BuildAttributes(e *spec.Element) func() (*template.Template, ImportSet) {
	return func() (*template.Template, ImportSet) {
		imports := ImportSet{}

		addTypes := &bytes.Buffer{}
		buf := &bytes.Buffer{}

		for _, attr := range e.Attributes {
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

				buf.WriteString(`
func (e *` + goIdent(e.Tag) + `Element) ` + goIdent(a.Name) + `(a ` + typeName + `) *` + goIdent(e.Tag) + `Element {
	e.attributes.Set("` + a.Name + `", string(a))
	
	return e
}
`)
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
