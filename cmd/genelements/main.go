package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/go-htemel/spec"

	"github.com/go-htemel/htemel/internal/generators"
)

func generate(pkg string, e *spec.Element) error {
	path, err := filepath.Abs(filepath.Join(pkg, e.Tag+".go"))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	var funcs = []generators.TemplateFunc{
		generators.BaseStruct,
		generators.BaseFunc,
		generators.BaseCondFunc,
		generators.BaseTernaryFunc,
		generators.ChildrenFunc,
		generators.WithFunc,
		generators.TextFunc,
		generators.TextfFunc,
		generators.IfFunc,
		generators.IfElseFunc,
		generators.IfThenFunc,
		generators.AddClassFunc,
		generators.ToggleClassFunc,
		generators.AttrFunc,
		generators.BuildAttributes(e),
		generators.RenderFunc,
	}

	if err = generators.SourceHeader(
		buf,
		"html",
		e,
		funcs...,
	); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
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
			return err
		}

		return os.WriteFile(path, fixed, 0644)
	})
}

func main() {
	sp := spec.Spec{}

	specFile, err := os.ReadFile("specs/html.json")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(specFile, &sp); err != nil {
		panic(err)
	}

	for _, e := range sp.Elements {
		// Merge global attributes into element
		for _, a := range sp.Attributes {
			if !slices.ContainsFunc(e.Attributes, func(eAttr spec.Attribute) bool {
				return a.GetName() == eAttr.GetName()
			}) {
				e.Attributes = append(e.Attributes, a)
			}
		}

		if err = generate(strings.ToLower(sp.Name), e); err != nil {
			panic(err)
		}
	}

	err = fmtFiles("html")
	if err != nil {
		panic(err)
	}
}
