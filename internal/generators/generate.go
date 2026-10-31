package generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/go-htemel/spec"
)

// FileName is the output file for an element. It is lower-cased so that the generated package never
// depends on a case-sensitive file system for tags with interior capitals (clipPath, feGaussianBlur).
func FileName(e *spec.Element) string {
	return strings.ToLower(e.Tag) + ".go"
}

// Generate writes one Go file per element of sp into dir, as package pkg. Global attributes are merged
// into every element first (an element's own attribute of the same name wins), and the resulting Go
// identifiers and file names are checked for collisions before anything is written.
func Generate(sp *spec.Spec, dir, pkg string) error {
	for _, e := range sp.Elements {
		for _, a := range sp.Attributes {
			if !slices.ContainsFunc(e.Attributes, func(eAttr spec.Attribute) bool {
				return a.GetName() == eAttr.GetName()
			}) {
				e.Attributes = append(e.Attributes, a)
			}
		}
	}

	if err := CheckNames(sp.Elements); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, e := range sp.Elements {
		if err := generateElement(dir, pkg, e); err != nil {
			return fmt.Errorf("<%s>: %w", e.Tag, err)
		}
	}

	return fmtFiles(dir)
}

func generateElement(dir, pkg string, e *spec.Element) error {
	buf := &bytes.Buffer{}

	funcs := []TemplateFunc{
		BaseStruct,
		BaseFunc,
		BaseCondFunc,
		BaseTernaryFunc,
		ChildrenFunc,
		WithFunc,
		TextFunc,
		TextfFunc,
		IfFunc,
		IfElseFunc,
		IfThenFunc,
		AddClassFunc,
		ToggleClassFunc,
		AttrFunc,
		BuildAttributes(e),
		RenderFunc,
	}

	if err := SourceHeader(buf, pkg, e, funcs...); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, FileName(e)), buf.Bytes(), 0o644)
}

func fmtFiles(dir string) error {
	opt := &imports.Options{
		AllErrors:  true,
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: false,
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		f, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fixed, err := imports.Process(path, f, opt)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		return os.WriteFile(path, fixed, 0o644)
	})
}

// elementMethods are the methods every generated element has regardless of its attributes; an
// attribute whose setter name equals one of these would not compile.
var elementMethods = []string{"With", "AddClass", "ToggleClass", "Attr", "AttrUnsafe", "BoolAttr", "Render"}

// containerMethods are the extra methods of non-void elements.
var containerMethods = []string{"Children", "Text", "Textf", "If", "IfElse", "IfThen"}

// CheckNames reports every Go identifier or output file name that the given elements (with globals
// already merged) would generate more than once, e.g. two tags differing only in case, two attributes
// that mangle to the same setter (stroke-width and strokeWidth), an attribute setter clashing with a
// built-in method, or an enum type clashing with another element's constructor.
func CheckNames(elements []*spec.Element) error {
	var problems []string

	pkgNames := map[string]string{} // identifier -> what declared it
	files := map[string]string{}    // file name -> tag

	declare := func(name, what string) {
		if prev, ok := pkgNames[name]; ok {
			problems = append(problems, fmt.Sprintf("package-level %s is declared by both %s and %s", name, prev, what))
			return
		}
		pkgNames[name] = what
	}

	for _, e := range elements {
		tag := goIdent(e.Tag)
		where := "<" + e.Tag + ">"

		if prev, ok := files[FileName(e)]; ok {
			problems = append(problems, fmt.Sprintf("file %s is produced by both <%s> and %s", FileName(e), prev, where))
		} else {
			files[FileName(e)] = e.Tag
		}

		declare(tag, where)
		declare(tag+"If", where)
		declare(tag+"Element", where)
		if !e.Void {
			declare(tag+"Ternary", where)
		}

		methods := map[string]string{}
		method := func(name, what string) {
			if prev, ok := methods[name]; ok {
				problems = append(problems, fmt.Sprintf("%s: method %s is declared by both %s and %s", where, name, prev, what))
				return
			}
			methods[name] = what
		}
		for _, m := range elementMethods {
			method(m, "the built-in "+m)
		}
		if !e.Void {
			for _, m := range containerMethods {
				method(m, "the built-in "+m)
			}
		}

		for _, attr := range e.Attributes {
			name := attr.GetName()
			what := fmt.Sprintf("attribute %q", name)
			method(goIdent(name), what)

			switch a := attr.(type) {
			case *spec.AttributeTypePrefixedCustom:
				method(goIdent(name)+"Unsafe", what)
			case *spec.AttributeTypeEnum:
				typeName := tag + goIdent(name)
				ewhere := where + " " + what
				declare(typeName, ewhere)
				for v := range a.Allowed {
					declare(typeName+goIdent(v), fmt.Sprintf("%s value %q", ewhere, v))
				}
				if a.AllowEmpty {
					declare(typeName+"Empty", ewhere+" (empty value)")
				}
				if a.AllowCustom {
					declare(typeName+"Custom", ewhere+" (custom value)")
				}
			}
		}
	}

	if len(problems) > 0 {
		slices.Sort(problems)
		return fmt.Errorf("generated names collide:\n  %s", strings.Join(problems, "\n  "))
	}

	return nil
}
