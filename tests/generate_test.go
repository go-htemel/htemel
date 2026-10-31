package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGeneratedCodeIsCurrent runs the generator into a temp dir and compares the output
// file-by-file with the committed html/ package.
func TestGeneratedCodeIsCurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generator comparison in -short mode")
	}

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	out := t.TempDir()
	cmd := exec.Command("go", "run", "./cmd/genelements",
		"-spec", filepath.Join(root, "specs", "html.json"),
		"-out", out)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generator failed: %v\n%s", err, b)
	}

	const hint = "run `go generate ./...` and commit the result"

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
		got, err := os.ReadFile(filepath.Join(root, "html", name))
		if err != nil {
			t.Errorf("html/%s is missing from the repository; %s", name, hint)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("html/%s is out of date with specs/html.json; %s", name, hint)
		}
	}

	committed, err := os.ReadDir(filepath.Join(root, "html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range committed {
		if filepath.Ext(e.Name()) == ".go" && !generated[e.Name()] {
			t.Errorf("html/%s is not produced by the generator (stale file?); %s", e.Name(), hint)
		}
	}
}
