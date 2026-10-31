package datastar

import (
	"strconv"
	"strings"
	"time"
)

// Modifier is a Datastar attribute modifier, rendered after the attribute name as __name followed
// by its arguments, each after a dot (for example __debounce.500ms.leading).
//
// Modifiers are values: the methods that add arguments return a new Modifier.
type Modifier struct {
	name string
	args []string
}

// Mod returns a modifier with the given name and arguments, for modifiers without a dedicated
// helper. Mod("debounce", "500ms", "leading") renders __debounce.500ms.leading.
func Mod(name string, args ...string) Modifier {
	return Modifier{name: name, args: args}
}

// With returns a copy of m with args appended.
func (m Modifier) With(args ...string) Modifier {
	all := make([]string, 0, len(m.args)+len(args))
	all = append(all, m.args...)
	all = append(all, args...)
	return Modifier{name: m.name, args: all}
}

// Leading adds the leading argument (fire on the leading edge), used with Debounce and Duration.
func (m Modifier) Leading() Modifier { return m.With("leading") }

// NoLeading adds the noleading argument (do not fire on the leading edge), used with Throttle.
func (m Modifier) NoLeading() Modifier { return m.With("noleading") }

// Trailing adds the trailing argument (also fire on the trailing edge), used with Throttle.
func (m Modifier) Trailing() Modifier { return m.With("trailing") }

// NoTrailing adds the notrailing argument (do not fire on the trailing edge), used with Debounce.
func (m Modifier) NoTrailing() Modifier { return m.With("notrailing") }

// String returns the modifier as it appears in an attribute name, including the leading "__".
func (m Modifier) String() string {
	if m.name == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("__")
	b.WriteString(m.name)
	for _, a := range m.args {
		b.WriteByte('.')
		b.WriteString(a)
	}
	return b.String()
}

// ms formats d in whole milliseconds, the form Datastar's time arguments accept.
func ms(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
}

// Event modifiers (data-on:*).

// Once runs the handler only once (data-on, data-on-intersect).
func Once() Modifier { return Mod("once") }

// Passive registers a passive event listener.
func Passive() Modifier { return Mod("passive") }

// Capture registers a capturing event listener.
func Capture() Modifier { return Mod("capture") }

// Prevent calls preventDefault on the event.
func Prevent() Modifier { return Mod("prevent") }

// Stop calls stopPropagation on the event.
func Stop() Modifier { return Mod("stop") }

// Window attaches the listener to the window instead of the element.
func Window() Modifier { return Mod("window") }

// Outside fires the handler when the event happens outside the element.
func Outside() Modifier { return Mod("outside") }

// ViewTransition wraps the expression in document.startViewTransition() where supported.
func ViewTransition() Modifier { return Mod("viewtransition") }

// Delay delays running the expression by d.
func Delay(d time.Duration) Modifier { return Mod("delay", ms(d)) }

// Debounce debounces the handler by d. Chain Leading or NoTrailing to change which edges fire.
func Debounce(d time.Duration) Modifier { return Mod("debounce", ms(d)) }

// Throttle throttles the handler to once per d. Chain NoLeading or Trailing to change which edges
// fire.
func Throttle(d time.Duration) Modifier { return Mod("throttle", ms(d)) }

// Duration sets the interval of data-on-interval (default one second). Chain Leading to also run
// the expression immediately.
func Duration(d time.Duration) Modifier { return Mod("duration", ms(d)) }

// Half triggers data-on-intersect when half of the element is visible.
func Half() Modifier { return Mod("half") }

// Full triggers data-on-intersect when the whole element is visible.
func Full() Modifier { return Mod("full") }

// Other modifiers.

// Case is a key casing understood by the __case modifier.
type Case string

// Key casings for CaseMod.
const (
	CaseCamel  Case = "camel"
	CaseKebab  Case = "kebab"
	CaseSnake  Case = "snake"
	CasePascal Case = "pascal"
)

// CaseMod converts the attribute key to the given casing (the __case modifier), for example to
// listen for a camelCase custom event or to create a snake_case signal.
func CaseMod(c Case) Modifier { return Mod("case", string(c)) }

// IfMissing only sets signals that do not exist yet (data-signals).
func IfMissing() Modifier { return Mod("ifmissing") }

// Terse renders data-json-signals output without indentation.
func Terse() Modifier { return Mod("terse") }

// Self makes data-ignore ignore only the element itself, not its descendants.
func Self() Modifier { return Mod("self") }

// Session makes data-persist use sessionStorage instead of localStorage.
func Session() Modifier { return Mod("session") }
