package htemel

import (
	"io"
	"iter"
)

// nilNode is a Node that renders nothing. It is returned by helpers instead of a nil interface,
// which would panic when rendered.
type nilNode struct{}

func (nilNode) Render(io.Writer) error { return nil }

// renderNode renders n, treating a nil Node as empty.
func renderNode(w io.Writer, n Node) error {
	if n == nil {
		return nil
	}
	return n.Render(w)
}

type mapNode[T any] struct {
	items []T
	fn    func(int, T) Node
}

func (m mapNode[T]) Render(w io.Writer) error {
	for i, item := range m.items {
		if err := renderNode(w, m.fn(i, item)); err != nil {
			return err
		}
	}
	return nil
}

type seqNode[T any] struct {
	seq iter.Seq[T]
	fn  func(T) Node
}

func (s seqNode[T]) Render(w io.Writer) error {
	var err error
	s.seq(func(v T) bool {
		err = renderNode(w, s.fn(v))
		return err == nil
	})
	return err
}

type seq2Node[K, V any] struct {
	seq iter.Seq2[K, V]
	fn  func(K, V) Node
}

func (s seq2Node[K, V]) Render(w io.Writer) error {
	var err error
	s.seq(func(k K, v V) bool {
		err = renderNode(w, s.fn(k, v))
		return err == nil
	})
	return err
}

// Each returns a Node that renders fn(item) for each item, in order. The slice is walked and fn
// is called at render time, so no intermediate slice of nodes is built. Nil results are skipped.
//
//	html.Ul().Children(
//		htemel.Each(names, func(n string) htemel.Node { return html.Li().Text(n) }),
//	)
func Each[T any](items []T, fn func(T) Node) Node {
	return mapNode[T]{items: items, fn: func(_ int, v T) Node { return fn(v) }}
}

// EachIndex is like Each but also passes the index of each item to fn.
//
//	htemel.EachIndex(rows, func(i int, r Row) htemel.Node { return html.Tr().Class(rowClass(i)) })
func EachIndex[T any](items []T, fn func(int, T) Node) Node {
	return mapNode[T]{items: items, fn: fn}
}

// ForEach returns a Node that ranges over seq at render time and renders fn(v) for each value.
// Nil results are skipped. It accepts anything that yields an iter.Seq, such as slices.Values or
// maps.Keys.
//
//	htemel.ForEach(slices.Values(names), func(n string) htemel.Node { return html.Li().Text(n) })
func ForEach[T any](seq iter.Seq[T], fn func(T) Node) Node {
	return seqNode[T]{seq: seq, fn: fn}
}

// ForEach2 is like ForEach for an iter.Seq2, such as maps.All or slices.All.
//
//	htemel.ForEach2(maps.All(m), func(k, v string) htemel.Node { return html.Dt().Text(k) })
func ForEach2[K, V any](seq iter.Seq2[K, V], fn func(K, V) Node) Node {
	return seq2Node[K, V]{seq: seq, fn: fn}
}

// If returns n when cond is true and a Node that renders nothing otherwise. Note that n is
// constructed by the caller either way; wrap construction in Each or a function if it is costly.
//
//	htemel.If(user != nil, html.P().Text("Welcome back"))
func If(cond bool, n Node) Node {
	if cond && n != nil {
		return n
	}
	return nilNode{}
}

// IfElse returns then when cond is true and otherwise when it is false. A nil choice renders
// nothing.
//
//	htemel.IfElse(loggedIn, html.A().Href("/logout").Text("Log out"), html.A().Href("/login").Text("Log in"))
func IfElse(cond bool, then, otherwise Node) Node {
	n := otherwise
	if cond {
		n = then
	}
	if n == nil {
		return nilNode{}
	}
	return n
}
