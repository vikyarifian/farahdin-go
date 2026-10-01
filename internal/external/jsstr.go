package external

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// ErrUnexpectedContent means an upstream page no longer has the shape the
// parsing rules expect. In the source app this surfaced as a TypeError on
// `undefined` (e.g. `.split(x)[1].trim()`), which was caught and swallowed.
var ErrUnexpectedContent = errors.New("unexpected upstream content")

// S is a JavaScript-flavoured string pipeline used to port the source app's
// scraping chains one call at a time. The first failing step poisons the
// chain, so a port reads like the original:
//
//	Str(body).Split("Nama:", 0).Replace("ARTI NAMA", "").TrimStart()
type S struct {
	v   string
	err error
}

// Str starts a pipeline.
func Str(s string) S { return S{v: s} }

// Value returns the result or the first error.
func (s S) Value() (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.v, nil
}

// Err returns the first error in the chain.
func (s S) Err() error { return s.err }

// String returns the current value (empty when the chain failed).
func (s S) String() string { return s.v }

// Split ports s.split(sep)[i]. An index past the end fails the chain,
// because the source app would then call a method on undefined.
func (s S) Split(sep string, i int) S {
	if s.err != nil {
		return s
	}
	parts := strings.Split(s.v, sep)
	if i < 0 || i >= len(parts) {
		return S{err: ErrUnexpectedContent}
	}
	return S{v: parts[i]}
}

// Replace ports s.replace("literal", repl): first occurrence only.
func (s S) Replace(old, repl string) S {
	if s.err != nil {
		return s
	}
	return S{v: strings.Replace(s.v, old, repl, 1)}
}

// ReplaceAll ports s.replaceAll("literal", repl).
func (s S) ReplaceAll(old, repl string) S {
	if s.err != nil {
		return s
	}
	return S{v: strings.ReplaceAll(s.v, old, repl)}
}

// ReplaceRe ports s.replace(/re/, repl) without the g flag: first match only.
// repl is literal.
func (s S) ReplaceRe(re *regexp.Regexp, repl string) S {
	if s.err != nil {
		return s
	}
	loc := re.FindStringIndex(s.v)
	if loc == nil {
		return s
	}
	return S{v: s.v[:loc[0]] + repl + s.v[loc[1]:]}
}

// ReplaceReAll ports s.replace(/re/g, repl). repl is literal.
func (s S) ReplaceReAll(re *regexp.Regexp, repl string) S {
	if s.err != nil {
		return s
	}
	return S{v: re.ReplaceAllLiteralString(s.v, repl)}
}

// TrimStart ports s.trimStart().
func (s S) TrimStart() S {
	if s.err != nil {
		return s
	}
	return S{v: strings.TrimLeftFunc(s.v, isJSSpace)}
}

// TrimEnd ports s.trimEnd().
func (s S) TrimEnd() S {
	if s.err != nil {
		return s
	}
	return S{v: strings.TrimRightFunc(s.v, isJSSpace)}
}

// Lines ports s.split('\n') as the final step of a chain.
func (s S) Lines() ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return strings.Split(s.v, "\n"), nil
}

// SplitAll ports s.split(sep) as the final step of a chain.
func (s S) SplitAll(sep string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return strings.Split(s.v, sep), nil
}

// isJSSpace matches JavaScript's whitespace + line terminator set used by trim().
func isJSSpace(r rune) bool {
	return unicode.IsSpace(r) || r == 0xFEFF
}

// jsSpace is the JavaScript `\s` class for use inside Go regular expressions.
const jsSpace = `[\t\n\v\f\r \x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]`

// Regular expressions shared by the ports.
var (
	// ReBlankLines ports /^\s*\n/gm.
	ReBlankLines = regexp.MustCompile(`(?m)^` + jsSpace + `*\n`)
	// ReHitungKembali ports /< Hitung Kembali.*$/s.
	ReHitungKembali = regexp.MustCompile(`(?s)< Hitung Kembali.*$`)
	// ReNewline ports /\n/ and /\n/gi.
	ReNewline = regexp.MustCompile(`\n`)
)
