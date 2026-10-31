package htemel

import (
	"io"
	"iter"
)

// renderNode renders n, treating a nil Node as empty.
func renderNode(w io.Writer, n Node) error {
	if n == nil {
		return nil
	}
	return n.Render(w)
}

// Each returns a Node that renders fn(item) for each item, in order. The slice is walked and fn
// is called at render time, so no intermediate slice of nodes is built. Nil results are skipped.
//
// The result is a NodeFunc, which is accepted as a child of any element: the nodes fn returns are
// not checked against the parent's content model.
//
//	html.Ul().Children(
//		htemel.Each(names, func(n string) htemel.Node { return html.Li().Text(n) }),
//	)
func Each[T any](items []T, fn func(T) Node) NodeFunc {
	return func(w io.Writer) error {
		for _, item := range items {
			if err := renderNode(w, fn(item)); err != nil {
				return err
			}
		}
		return nil
	}
}

// EachIndex is like Each but also passes the index of each item to fn.
//
//	htemel.EachIndex(rows, func(i int, r Row) htemel.Node { return html.Tr().Class(rowClass(i)) })
func EachIndex[T any](items []T, fn func(int, T) Node) NodeFunc {
	return func(w io.Writer) error {
		for i, item := range items {
			if err := renderNode(w, fn(i, item)); err != nil {
				return err
			}
		}
		return nil
	}
}

// ForEach returns a Node that ranges over seq at render time and renders fn(v) for each value.
// Nil results are skipped. It accepts anything that yields an iter.Seq, such as slices.Values or
// maps.Keys.
//
//	htemel.ForEach(slices.Values(names), func(n string) htemel.Node { return html.Li().Text(n) })
func ForEach[T any](seq iter.Seq[T], fn func(T) Node) NodeFunc {
	return func(w io.Writer) error {
		var err error
		seq(func(v T) bool {
			err = renderNode(w, fn(v))
			return err == nil
		})
		return err
	}
}

// ForEach2 is like ForEach for an iter.Seq2, such as maps.All or slices.All.
//
//	htemel.ForEach2(maps.All(m), func(k, v string) htemel.Node { return html.Dt().Text(k) })
func ForEach2[K, V any](seq iter.Seq2[K, V], fn func(K, V) Node) NodeFunc {
	return func(w io.Writer) error {
		var err error
		seq(func(k K, v V) bool {
			err = renderNode(w, fn(k, v))
			return err == nil
		})
		return err
	}
}

// If returns n when cond is true and a Node that renders nothing otherwise. Like Each, the result
// is a NodeFunc and is accepted as a child of any element. Note that n is
// constructed by the caller either way; wrap construction in Each or a function if it is costly.
//
//	htemel.If(user != nil, html.P().Text("Welcome back"))
func If(cond bool, n Node) NodeFunc {
	if !cond {
		return nop
	}
	return Unchecked(n)
}

// IfElse returns then when cond is true and otherwise when it is false. A nil choice renders
// nothing.
//
//	htemel.IfElse(loggedIn, html.A().Href("/logout").Text("Log out"), html.A().Href("/login").Text("Log in"))
func IfElse(cond bool, then, otherwise Node) NodeFunc {
	if cond {
		return Unchecked(then)
	}
	return Unchecked(otherwise)
}
