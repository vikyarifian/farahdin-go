package external

import (
	"strings"

	"golang.org/x/net/html"
)

// Doc is a parsed HTML document with the small subset of cheerio behaviour
// the source app relied on: CSS lookup by tag/#id/.class with child (>) and
// descendant combinators, .text() and .html().
type Doc struct {
	root *html.Node
}

// ParseHTML parses an HTML document the same way browsers (and parse5,
// which cheerio used) build the tree.
func ParseHTML(s string) (*Doc, error) {
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return nil, err
	}
	return &Doc{root: n}, nil
}

// Text ports `$.text()` on the loaded root: the text content of the whole document.
func (d *Doc) Text() string {
	var b strings.Builder
	textContent(d.root, &b)
	return b.String()
}

// Find returns every element matching sel, in document order.
func (d *Doc) Find(sel string) Selection {
	steps := parseSelector(sel)
	var out Selection
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && matchSteps(n, steps, len(steps)-1) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(d.root)
	return out
}

// Selection is a set of matched element nodes.
type Selection []*html.Node

// Text ports cheerio's .text(): concatenated text content of every match,
// including <script> contents (the source app stripped those by hand).
func (s Selection) Text() string {
	var b strings.Builder
	for _, n := range s {
		textContent(n, &b)
	}
	return b.String()
}

// HTML ports cheerio's .html(): inner HTML of the first match, serialized
// like parse5 (void elements without a slash, &lt; escaping in text).
// ok is false when nothing matched (cheerio returns null).
func (s Selection) HTML() (string, bool) {
	if len(s) == 0 {
		return "", false
	}
	var b strings.Builder
	for c := s[0].FirstChild; c != nil; c = c.NextSibling {
		serialize(c, &b)
	}
	return b.String(), true
}

// Attr ports .attr(name) on the first match.
func (s Selection) Attr(name string) (string, bool) {
	if len(s) == 0 {
		return "", false
	}
	for _, a := range s[0].Attr {
		if a.Key == name {
			return a.Val, true
		}
	}
	return "", false
}

func textContent(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		b.WriteString(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		textContent(c, b)
	}
}

// --- selectors ---

type compound struct {
	tag     string
	id      string
	classes []string
	// child is true when this step must be a direct child of the previous one.
	child bool
}

func parseSelector(sel string) []compound {
	var steps []compound
	child := false
	for _, tok := range strings.Fields(strings.ReplaceAll(sel, ">", " > ")) {
		if tok == ">" {
			child = true
			continue
		}
		c := compound{child: child}
		child = false
		rest := tok
		// tag is everything before the first # or .
		if i := strings.IndexAny(rest, "#."); i != 0 {
			if i < 0 {
				c.tag, rest = rest, ""
			} else {
				c.tag, rest = rest[:i], rest[i:]
			}
		}
		for rest != "" {
			kind := rest[0]
			rest = rest[1:]
			end := strings.IndexAny(rest, "#.")
			if end < 0 {
				end = len(rest)
			}
			name := rest[:end]
			rest = rest[end:]
			if kind == '#' {
				c.id = name
			} else {
				c.classes = append(c.classes, name)
			}
		}
		steps = append(steps, c)
	}
	return steps
}

func matchSteps(n *html.Node, steps []compound, i int) bool {
	if i < 0 {
		return true
	}
	if !matchCompound(n, steps[i]) {
		return false
	}
	if i == 0 {
		return true
	}
	if steps[i].child {
		p := n.Parent
		return p != nil && p.Type == html.ElementNode && matchSteps(p, steps, i-1)
	}
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && matchSteps(p, steps, i-1) {
			return true
		}
	}
	return false
}

func matchCompound(n *html.Node, c compound) bool {
	if c.tag != "" && c.tag != "*" && n.Data != c.tag {
		return false
	}
	if c.id != "" && attr(n, "id") != c.id {
		return false
	}
	if len(c.classes) > 0 {
		have := strings.Fields(attr(n, "class"))
		for _, want := range c.classes {
			found := false
			for _, h := range have {
				if h == want {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// --- parse5-compatible serializer ---

var voidElements = map[string]bool{
	"area": true, "base": true, "basefont": true, "bgsound": true, "br": true, "col": true,
	"embed": true, "frame": true, "hr": true, "img": true, "input": true, "keygen": true,
	"link": true, "meta": true, "param": true, "source": true, "track": true, "wbr": true,
}

var rawTextParents = map[string]bool{
	"style": true, "script": true, "xmp": true, "iframe": true, "noembed": true,
	"noframes": true, "plaintext": true, "noscript": true,
}

// nbsp is U+00A0, which parse5 serializes as &nbsp;.
var nbsp = string(rune(0x00A0))

var (
	textEscaper = strings.NewReplacer("&", "&amp;", nbsp, "&nbsp;", "<", "&lt;", ">", "&gt;")
	attrEscaper = strings.NewReplacer("&", "&amp;", nbsp, "&nbsp;", `"`, "&quot;")
)

func serialize(n *html.Node, b *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		if n.Parent != nil && n.Parent.Type == html.ElementNode && rawTextParents[n.Parent.Data] {
			b.WriteString(n.Data)
		} else {
			b.WriteString(textEscaper.Replace(n.Data))
		}
	case html.CommentNode:
		b.WriteString("<!--" + n.Data + "-->")
	case html.ElementNode:
		b.WriteString("<" + n.Data)
		for _, a := range n.Attr {
			name := a.Key
			if a.Namespace != "" {
				name = a.Namespace + ":" + a.Key
			}
			b.WriteString(" " + name + `="` + attrEscaper.Replace(a.Val) + `"`)
		}
		b.WriteString(">")
		if voidElements[n.Data] {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			serialize(c, b)
		}
		b.WriteString("</" + n.Data + ">")
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			serialize(c, b)
		}
	}
}
