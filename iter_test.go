package htemel_test

import (
	"errors"
	"io"
	"maps"
	"slices"
	"testing"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"
)

func li(s string) htemel.Node { return html.Li().Text(s) }

func render(t *testing.T, n htemel.Node) string {
	t.Helper()
	s, err := htemel.RenderString(n)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEach(t *testing.T) {
	got := render(t, htemel.Each([]string{"a", "<b>"}, li))
	if want := "<li>a</li><li>&lt;b&gt;</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if got := render(t, htemel.Each[string](nil, li)); got != "" {
		t.Errorf("empty: %q", got)
	}
	skip := htemel.Each([]int{1, 2, 3}, func(i int) htemel.Node {
		if i == 2 {
			return nil
		}
		return li(string(rune('0' + i)))
	})
	if got, want := render(t, skip), "<li>1</li><li>3</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestMapIndex(t *testing.T) {
	n := htemel.EachIndex([]string{"x", "y"}, func(i int, s string) htemel.Node {
		return html.Li().Text(s + string(rune('0'+i)))
	})
	if got, want := render(t, n), "<li>x0</li><li>y1</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestForEach(t *testing.T) {
	n := htemel.ForEach(slices.Values([]string{"a", "b"}), li)
	if got, want := render(t, n), "<li>a</li><li>b</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	// Rendering twice re-iterates.
	if again := render(t, n); again != "<li>a</li><li>b</li>" {
		t.Errorf("second render got %q", again)
	}
	m := map[string]string{"k": "v"}
	n2 := htemel.ForEach(maps.Keys(m), li)
	if got, want := render(t, n2), "<li>k</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestForEach2(t *testing.T) {
	n := htemel.ForEach2(slices.All([]string{"a", "b"}), func(i int, s string) htemel.Node {
		if i == 0 {
			return nil
		}
		return li(s)
	})
	if got, want := render(t, n), "<li>b</li>"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

type errNode struct{ err error }

func (e errNode) Render(io.Writer) error { return e.err }

func TestErrorStopsIteration(t *testing.T) {
	boom := errors.New("boom")
	calls := 0
	fn := func(s string) htemel.Node { calls++; return errNode{boom} }
	for name, n := range map[string]htemel.Node{
		"map":     htemel.Each([]string{"a", "b"}, fn),
		"foreach": htemel.ForEach(slices.Values([]string{"a", "b"}), fn),
	} {
		calls = 0
		if _, err := htemel.RenderString(n); !errors.Is(err, boom) {
			t.Errorf("%s: err = %v", name, err)
		}
		if calls != 1 {
			t.Errorf("%s: calls = %d", name, calls)
		}
	}
}

func TestIf(t *testing.T) {
	p := html.P().Text("x")
	if got := render(t, htemel.If(true, p)); got != "<p>x</p>" {
		t.Errorf("got %q", got)
	}
	if got := render(t, htemel.If(false, p)); got != "" {
		t.Errorf("got %q", got)
	}
	if got := render(t, htemel.If(true, nil)); got != "" {
		t.Errorf("nil: %q", got)
	}
}

func TestIfElse(t *testing.T) {
	a, b := html.P().Text("a"), html.P().Text("b")
	if got := render(t, htemel.IfElse(true, a, b)); got != "<p>a</p>" {
		t.Errorf("got %q", got)
	}
	if got := render(t, htemel.IfElse(false, a, b)); got != "<p>b</p>" {
		t.Errorf("got %q", got)
	}
	if got := render(t, htemel.IfElse(false, a, nil)); got != "" {
		t.Errorf("nil: %q", got)
	}
}
