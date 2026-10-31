package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// generatedPackages lists every package produced by cmd/genelements (see the go:generate lines in
// htemel.go): the spec it is generated from and the committed directory it must match.
var generatedPackages = []struct{ spec, dir string }{
	{spec: "html.json", dir: "html"},
	{spec: "svg.json", dir: "svg"},
	{spec: "mathml.json", dir: "mathml"},
}

// TestGeneratedCodeIsCurrent runs the generator for each spec into a temp dir and compares the output
// file-by-file with the committed package directory, and the generated slots_<dir>.go file of
// package htemel with the committed one.
func TestGeneratedCodeIsCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator comparison in -short mode")
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range generatedPackages {
		t.Run(p.dir, func(t *testing.T) {
			checkGenerated(t, root, p.spec, p.dir)
		})
	}
}

func checkGenerated(t *testing.T, root, specName, dir string) {
	out := t.TempDir()
	slots := filepath.Join(t.TempDir(), "slots_"+dir+".go")
	cmd := exec.Command("go", "run", "./cmd/genelements",
		"-spec", filepath.Join(root, "specs", specName),
		"-out", out,
		"-pkg", dir,
		"-slots", slots)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generator failed: %v\n%s", err, b)
	}

	const hint = "run `go generate ./...` and commit the result"

	if want, err := os.ReadFile(slots); err != nil {
		t.Fatal(err)
	} else if got, err := os.ReadFile(filepath.Join(root, "slots_"+dir+".go")); err != nil || string(got) != string(want) {
		t.Errorf("slots_%s.go is missing or out of date with specs/%s; %s", dir, specName, hint)
	}

	generated := map[string]bool{}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		generated[name] = true

		want, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(root, dir, name))
		if err != nil {
			t.Errorf("%s/%s is missing from the repository; %s", dir, name, hint)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("%s/%s is out of date with specs/%s; %s", dir, name, specName, hint)
		}
	}

	committed, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range committed {
		if filepath.Ext(e.Name()) == ".go" && !generated[e.Name()] {
			t.Errorf("%s/%s is not produced by the generator (stale file?); %s", dir, e.Name(), hint)
		}
	}
}
