// Command genelements generates a Go package of element builders from a go-htemel/spec JSON file.
//
//	go run ./cmd/genelements -spec specs/html.json            # writes ./html, package html
//	go run ./cmd/genelements -spec specs/svg.json -out svg    # writes ./svg,  package svg
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"os"
	"strings"

	"github.com/go-htemel/spec"

	"github.com/go-htemel/htemel/internal/generators"
)

func main() {
	specPath := flag.String("spec", "specs/html.json", "path to the element spec JSON file")
	outDir := flag.String("out", "", "output directory for generated files (default: lower-cased spec name, e.g. html)")
	pkgName := flag.String("pkg", "", "Go package name of the generated files (default: lower-cased spec name, e.g. html)")
	flag.Parse()

	if err := run(*specPath, *outDir, *pkgName); err != nil {
		fmt.Fprintln(os.Stderr, "genelements:", err)
		os.Exit(1)
	}
}

func run(specPath, out, pkg string) error {
	specFile, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	sp := spec.Spec{}
	if err = json.Unmarshal(specFile, &sp); err != nil {
		return fmt.Errorf("%s: %w", specPath, err)
	}

	if out == "" {
		out = strings.ToLower(sp.Name)
	}
	if pkg == "" {
		pkg = strings.ToLower(sp.Name)
	}
	if !token.IsIdentifier(pkg) || token.IsKeyword(pkg) {
		return fmt.Errorf("invalid Go package name %q (set -pkg)", pkg)
	}

	return generators.Generate(&sp, out, pkg)
}
